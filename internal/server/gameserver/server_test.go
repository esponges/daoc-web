package gameserver

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"daocweb/internal/server/auth"
	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
	"daocweb/internal/server/world"
)

// testServer runs the whole server on the real world store and converted
// zones, skipping when this machine has neither.
func testServer(t *testing.T) (*httptest.Server, *http.Client, *world.World) {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	worldDB := filepath.Join(root, "server", "data", "world.db")
	webDir := filepath.Join(root, "web")
	if _, err := os.Stat(worldDB); err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	if _, err := os.Stat(filepath.Join(webDir, "data", "zone105", "zone.json")); err != nil {
		t.Skip("zone 105 is not converted")
	}
	cat, err := world.LoadCatalog(worldDB)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	a, err := auth.New(auth.Config{AdminPassword: "test-password", SessionKey: []byte("0123456789abcdef0123456789abcdef")}, st)
	if err != nil {
		t.Fatal(err)
	}
	w := world.New(world.Config{WebData: filepath.Join(webDir, "data"), Cat: cat, Store: st, Log: t.Logf})
	srv := &Server{Auth: a, World: w, Store: st, WebDir: webDir, Log: t.Logf}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(func() { ts.Close(); w.Shutdown(); st.Close() })
	jar, _ := cookiejar.New(nil)
	return ts, &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, w
}

func post(t *testing.T, c *http.Client, url string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	r, err := c.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func login(t *testing.T, ts *httptest.Server, c *http.Client) {
	t.Helper()
	if r := post(t, c, ts.URL+"/auth/admin", map[string]string{"password": "test-password"}); r.StatusCode != 204 {
		t.Fatalf("login: %d", r.StatusCode)
	}
}

func createChar(t *testing.T, ts *httptest.Server, c *http.Client, name string) int64 {
	t.Helper()
	r := post(t, c, ts.URL+"/api/characters", map[string]string{"name": name, "arch": "troll-warrior"})
	if r.StatusCode != 201 {
		t.Fatalf("create %s: %d", name, r.StatusCode)
	}
	var v struct {
		ID   int64
		Zone int
	}
	json.NewDecoder(r.Body).Decode(&v)
	return v.ID
}

type session struct {
	t    *testing.T
	conn *websocket.Conn
}

func connect(t *testing.T, ts *httptest.Server, c *http.Client, char int64) *session {
	t.Helper()
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	h := http.Header{}
	for _, ck := range c.Jar.Cookies(mustURL(ts.URL)) {
		h.Add("Cookie", ck.String())
	}
	conn, _, err := websocket.DefaultDialer.Dial(u, h)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	s := &session{t: t, conn: conn}
	s.send(protocol.Hello{T: protocol.THello, V: protocol.Version, Char: char})
	return s
}

func mustURL(s string) *urlT { u, _ := parseURL(s); return u }

func (s *session) send(v any) {
	b, _ := json.Marshal(v)
	s.conn.WriteMessage(websocket.TextMessage, b)
}

// until reads messages until one of type typ arrives.
func (s *session) until(typ string, within time.Duration) []byte {
	s.t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		s.conn.SetReadDeadline(deadline)
		_, b, err := s.conn.ReadMessage()
		if err != nil {
			s.t.Fatalf("waiting for %s: %v", typ, err)
		}
		var e protocol.Envelope
		json.Unmarshal(b, &e)
		if e.T == typ {
			return b
		}
		if e.T == protocol.TErr {
			s.t.Fatalf("server error while waiting for %s: %s", typ, b)
		}
	}
	s.t.Fatalf("no %s within %v", typ, within)
	return nil
}

func TestLoginRequired(t *testing.T) {
	ts, c, _ := testServer(t)
	// The game page redirects to the login page.
	r, _ := c.Get(ts.URL + "/")
	if r.StatusCode != 302 || r.Header.Get("Location") != "/login.html" {
		t.Errorf("/ without a session: %d %s", r.StatusCode, r.Header.Get("Location"))
	}
	// The socket and the API refuse an anonymous caller.
	if r, _ := c.Get(ts.URL + "/api/me"); r.StatusCode != 401 {
		t.Errorf("/api/me: %d", r.StatusCode)
	}
	_, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http")+"/ws", nil)
	if err == nil || resp == nil || resp.StatusCode != 401 {
		t.Errorf("anonymous websocket: %v %v", err, resp)
	}
	// A wrong password gets nothing.
	if r := post(t, c, ts.URL+"/auth/admin", map[string]string{"password": "nope"}); r.StatusCode != 401 {
		t.Errorf("wrong password: %d", r.StatusCode)
	}
}

func TestCreateCharacterAndPlay(t *testing.T) {
	ts, c, w := testServer(t)
	login(t, ts, c)
	id := createChar(t, ts, c, "Grimmtroll")
	if r := post(t, c, ts.URL+"/api/characters", map[string]string{"name": "grimmTROLL", "arch": "troll-warrior"}); r.StatusCode != 409 {
		t.Errorf("duplicate name: %d", r.StatusCode)
	}
	if r := post(t, c, ts.URL+"/api/characters", map[string]string{"name": "x1", "arch": "troll-warrior"}); r.StatusCode != 400 {
		t.Errorf("bad name: %d", r.StatusCode)
	}
	var me struct {
		Characters []struct {
			Zone     int
			ZoneName string
			Arch     string
		}
	}
	r, _ := c.Get(ts.URL + "/api/me")
	json.NewDecoder(r.Body).Decode(&me)
	if len(me.Characters) != 1 || me.Characters[0].Zone != 105 || me.Characters[0].Arch != "troll-warrior" {
		t.Fatalf("a new troll should start in Myrkwood Forest (105): %+v", me.Characters)
	}

	s := connect(t, ts, c, id)
	var wel protocol.Welcome
	json.Unmarshal(s.until(protocol.TWelcome, 5*time.Second), &wel)
	if wel.Zone != 105 || wel.Name != "Grimmtroll" || wel.Char != "troll-warrior" || wel.Z == 0 {
		t.Fatalf("welcome = %+v", wel)
	}
	var inv protocol.Inv
	json.Unmarshal(s.until(protocol.TInv, 5*time.Second), &inv)
	// The kit is the table's for class 22; a store from before the table gives none.
	if want := len(w.Cat.StartEquip[22]); len(inv.Items) != want {
		t.Errorf("a new warrior carries %d items, want the %d the starting-equipment table lists: %+v", len(inv.Items), want, inv.Items)
	}
	var snap protocol.Snap
	json.Unmarshal(s.until(protocol.TSnap, 5*time.Second), &snap)
	if len(snap.Ents) == 0 {
		t.Error("Myrkwood Forest has creatures near its start; the first snapshot shows none")
	}

	// Run north for a second: the server moves us about 191 units and echoes the ack.
	s.send(protocol.Input{T: protocol.TInput, Seq: 60, Steps: [][5]int{{60, 1, 0, 0, protocol.FlagRun}}})
	var last protocol.Snap
	for i := 0; i < 80; i++ {
		json.Unmarshal(s.until(protocol.TSnap, 5*time.Second), &last)
		if last.Ack >= 60 {
			break
		}
	}
	if last.Ack != 60 {
		t.Fatalf("ack = %d", last.Ack)
	}
	if d := math.Hypot(last.You.X-wel.X, last.You.Y-wel.Y); math.Abs(d-world.Run) > 3 {
		t.Errorf("moved %.1f units, want about %v", d, world.Run)
	}

	// The same character cannot be played twice at once.
	s2 := connect(t, ts, c, id)
	s2.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, b, _ := s2.conn.ReadMessage()
	if !strings.Contains(string(b), "already") {
		t.Errorf("second session got %s", b)
	}
	_ = w
}

func TestVersionMismatchRefused(t *testing.T) {
	ts, c, _ := testServer(t)
	login(t, ts, c)
	id := createChar(t, ts, c, "Oldclient")
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	h := http.Header{}
	for _, ck := range c.Jar.Cookies(mustURL(ts.URL)) {
		h.Add("Cookie", ck.String())
	}
	conn, _, err := websocket.DefaultDialer.Dial(u, h)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	b, _ := json.Marshal(protocol.Hello{T: protocol.THello, V: 0, Char: id})
	conn.WriteMessage(websocket.TextMessage, b)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, _ := conn.ReadMessage()
	if !strings.Contains(string(msg), `"code":"version"`) {
		t.Errorf("old client got %s", msg)
	}
}

func TestCharacterPersistsAcrossSessions(t *testing.T) {
	ts, c, _ := testServer(t)
	login(t, ts, c)
	id := createChar(t, ts, c, "Returner")
	s := connect(t, ts, c, id)
	var wel protocol.Welcome
	json.Unmarshal(s.until(protocol.TWelcome, 5*time.Second), &wel)
	s.send(protocol.Input{T: protocol.TInput, Seq: 30, Steps: [][5]int{{30, 1, 0, 0, protocol.FlagRun}}})
	var last protocol.Snap
	for i := 0; i < 80 && last.Ack < 30; i++ {
		json.Unmarshal(s.until(protocol.TSnap, 5*time.Second), &last)
	}
	s.conn.Close()
	// Wait for the server to notice and save, then come back.
	var wel2 protocol.Welcome
	for i := 0; i < 40; i++ {
		time.Sleep(100 * time.Millisecond)
		s2 := connect(t, ts, c, id)
		s2.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, b, err := s2.conn.ReadMessage()
		if err == nil && strings.Contains(string(b), `"t":"welcome"`) {
			json.Unmarshal(b, &wel2)
			break
		}
		s2.conn.Close()
	}
	if wel2.Name != "Returner" {
		t.Fatal("could not log back in")
	}
	if math.Hypot(wel2.X-last.You.X, wel2.Y-last.You.Y) > 5 {
		t.Errorf("came back at %.0f,%.0f, left at %.0f,%.0f", wel2.X, wel2.Y, last.You.X, last.You.Y)
	}
}

type urlT = url.URL

func parseURL(s string) (*url.URL, error) { return url.Parse(s) }

// The whole Discord path through the real server and a stand-in Discord:
// authorize redirect, callback, character, a play session.
func TestDiscordLoginToPlay(t *testing.T) {
	discord := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/oauth2/token":
			json.NewEncoder(w).Encode(map[string]string{"access_token": "tok"})
		case "/api/users/@me":
			json.NewEncoder(w).Encode(map[string]string{"id": "777", "username": "grimm", "global_name": "Grimm"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer discord.Close()
	root := filepath.Join("..", "..", "..")
	worldDB := filepath.Join(root, "server", "data", "world.db")
	if _, err := os.Stat(worldDB); err != nil {
		t.Skip("no world store")
	}
	cat, err := world.LoadCatalog(worldDB)
	if err != nil {
		t.Fatal(err)
	}
	st, _ := store.Open(":memory:")
	defer st.Close()
	var ts *httptest.Server
	srvHandler := http.NewServeMux()
	ts = httptest.NewServer(srvHandler)
	defer ts.Close()
	a, err := auth.New(auth.Config{DiscordID: "cid", DiscordSecret: "shh", DiscordBase: discord.URL, PublicURL: ts.URL,
		SessionKey: []byte("0123456789abcdef0123456789abcdef")}, st)
	if err != nil {
		t.Fatal(err)
	}
	w := world.New(world.Config{WebData: filepath.Join(root, "web", "data"), Cat: cat, Store: st, Log: t.Logf})
	defer w.Shutdown()
	srv := &Server{Auth: a, World: w, Store: st, WebDir: filepath.Join(root, "web"), Log: t.Logf}
	srvHandler.Handle("/", srv.Handler())
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

	// 1. start: redirected to Discord with our state cookie
	r, _ := c.Get(ts.URL + "/auth/discord")
	loc, _ := url.Parse(r.Header.Get("Location"))
	if r.StatusCode != 302 || !strings.HasPrefix(r.Header.Get("Location"), discord.URL+"/oauth2/authorize") {
		t.Fatalf("start: %d %s", r.StatusCode, r.Header.Get("Location"))
	}
	// 2. Discord sends the browser back with a code and our state
	r, _ = c.Get(ts.URL + "/auth/discord/callback?code=good&state=" + loc.Query().Get("state"))
	if r.StatusCode != 302 || r.Header.Get("Location") != "/login.html" {
		t.Fatalf("callback: %d %s", r.StatusCode, r.Header.Get("Location"))
	}
	// 3. logged in: an ordinary account, not an administrator
	var me struct {
		Name  string
		Admin bool
	}
	r, _ = c.Get(ts.URL + "/api/me")
	json.NewDecoder(r.Body).Decode(&me)
	if me.Name != "Grimm" || me.Admin {
		t.Fatalf("me = %+v", me)
	}
	// 4. a character, and a play session
	id := createChar(t, ts, c, "Discordtroll")
	s := connect(t, ts, c, id)
	var wel protocol.Welcome
	json.Unmarshal(s.until(protocol.TWelcome, 5*time.Second), &wel)
	if wel.Name != "Discordtroll" || wel.Admin || wel.Zone != 105 {
		t.Errorf("welcome = %+v", wel)
	}
	// 5. and not an administrator: a GM command is refused
	s.send(protocol.GM{T: protocol.TGM, Line: "goto 103"})
	var e protocol.Err
	json.Unmarshal(s.until(protocol.TErr, 5*time.Second), &e)
	if e.Code != "gm" {
		t.Errorf("a Discord player used a GM command: %+v", e)
	}
}

// Creation takes an optional stat build, checked by the rules: 30 points
// exactly, nothing lowered. Without one the class default is used.
func TestCreateCharacterWithBuild(t *testing.T) {
	ts, c, w := testServer(t)
	login(t, ts, c)
	type body struct {
		Name  string  `json:"name"`
		Arch  string  `json:"arch"`
		Build *[8]int `json:"build,omitempty"`
	}
	if r := post(t, c, ts.URL+"/api/characters", body{"Overspent", "troll-warrior", &[8]int{11, 10, 10}}); r.StatusCode != 400 {
		t.Errorf("a 31-point build: %d, want 400", r.StatusCode)
	}
	if r := post(t, c, ts.URL+"/api/characters", body{"Strongarm", "troll-warrior", &[8]int{15, 5, 0, 5}}); r.StatusCode != 201 {
		t.Errorf("a valid build: %d, want 201", r.StatusCode)
	}
	if r := post(t, c, ts.URL+"/api/characters", body{Name: "Plain", Arch: "troll-warrior"}); r.StatusCode != 201 {
		t.Errorf("no build: %d, want 201", r.StatusCode)
	}
	list, err := w.Store.Characters(1)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string][8]int{}
	for _, ch := range list {
		got[ch.Name] = ch.Stats
		if !ch.HasStats || ch.XPScheme != store.XPSchemeRules {
			t.Errorf("%s: stats %v scheme %d", ch.Name, ch.HasStats, ch.XPScheme)
		}
		if b := ch.Bind; b == nil || b.Zone != ch.Zone || b.X != ch.X || b.Y != ch.Y {
			t.Errorf("%s is bound at %+v, want their start", ch.Name, ch.Bind)
		}
	}
	if _, ok := got["Overspent"]; ok {
		t.Error("the refused build made a character")
	}
	if got["Strongarm"] != [8]int{115, 75, 35, 40, 60, 60, 60, 60} {
		t.Errorf("Strongarm %v", got["Strongarm"])
	}
	if got["Plain"] != [8]int{110, 80, 45, 35, 60, 60, 60, 60} {
		t.Errorf("Plain %v", got["Plain"])
	}
}
