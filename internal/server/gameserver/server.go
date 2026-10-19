// Package gameserver is the HTTP face of the game: login, the character API,
// the WebSocket that carries a play session, and the static viewer.
package gameserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"daocweb/internal/server/auth"
	"daocweb/internal/server/protocol"
	"daocweb/internal/server/store"
	"daocweb/internal/server/world"
)

// Server serves one game.
type Server struct {
	Auth   *auth.Auth
	World  *world.World
	Store  *store.Store
	WebDir string
	Log    func(format string, a ...any)
	up     websocket.Upgrader
}

// MaxCharacters is how many characters one account may keep.
const MaxCharacters = 4

// Handler builds the routes.
func (s *Server) Handler() http.Handler {
	if s.Log == nil {
		s.Log = func(string, ...any) {}
	}
	s.up = websocket.Upgrader{
		ReadBufferSize: 4096, WriteBufferSize: 16384,
		// Only pages served by this server may open a session.
		CheckOrigin: func(r *http.Request) bool {
			o := r.Header.Get("Origin")
			if o == "" {
				return true
			}
			u, err := url.Parse(o)
			return err == nil && u.Host == r.Host
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/config", s.config)
	mux.HandleFunc("/api/me", s.Auth.Require(s.me))
	mux.HandleFunc("/api/characters", s.Auth.Require(s.characters))
	mux.HandleFunc("/api/characters/", s.Auth.Require(s.character))
	mux.HandleFunc("/auth/admin", s.Auth.AdminLogin)
	mux.HandleFunc("/auth/discord", s.Auth.DiscordStart)
	mux.HandleFunc("/auth/discord/callback", s.Auth.DiscordCallback)
	mux.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		s.Auth.Logout(w)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/ws", s.ws)
	files := http.FileServer(http.Dir(s.WebDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// The game page needs a login; the login page and assets do not.
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			if _, ok := s.Auth.Session(r); !ok {
				http.Redirect(w, r, "/login.html", http.StatusFound)
				return
			}
		}
		w.Header().Set("Cache-Control", "no-cache")
		files.ServeHTTP(w, r)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (s *Server) config(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"discord": s.Auth.HasDiscord(), "admin": s.Auth.HasAdmin(), "protocol": protocol.Version})
}

type charView struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Arch     string `json:"arch"`
	Label    string `json:"label"`
	Level    int    `json:"level"`
	Zone     int    `json:"zone"`
	ZoneName string `json:"zoneName"`
}

func (s *Server) view(c *store.Character) charView {
	a := world.ArchetypeFor(c.Race, c.Class)
	zn := ""
	if zi := s.World.Cat.Zones[c.Zone]; zi != nil {
		zn = zi.Name
	}
	return charView{ID: c.ID, Name: c.Name, Arch: a.Key, Label: a.Label, Level: c.Level, Zone: c.Zone, ZoneName: zn}
}

func (s *Server) me(w http.ResponseWriter, r *http.Request, sess *auth.Session) {
	acct, err := s.Store.Account(sess.AccountID)
	if err != nil {
		http.Error(w, "unknown account", http.StatusUnauthorized)
		return
	}
	list, err := s.Store.Characters(acct.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	out := []charView{}
	for _, c := range list {
		out = append(out, s.view(c))
	}
	arch := []map[string]any{}
	for _, a := range world.Archetypes {
		arch = append(arch, map[string]any{"key": a.Key, "label": a.Label})
	}
	writeJSON(w, map[string]any{"name": acct.Name, "admin": sess.Admin, "characters": out, "archetypes": arch, "max": MaxCharacters})
}

func (s *Server) characters(w http.ResponseWriter, r *http.Request, sess *auth.Session) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name  string  `json:"name"`
		Arch  string  `json:"arch"`
		Build *[8]int `json:"build"` // how far each stat is raised; absent for the class default
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var arch *world.Archetype
	for i := range world.Archetypes {
		if world.Archetypes[i].Key == body.Arch {
			arch = &world.Archetypes[i]
		}
	}
	if arch == nil {
		http.Error(w, "choose a character type", http.StatusBadRequest)
		return
	}
	list, _ := s.Store.Characters(sess.AccountID)
	if len(list) >= MaxCharacters {
		http.Error(w, "you have the most characters allowed", http.StatusConflict)
		return
	}
	stats, err := world.CreationStats(arch.Race, arch.Class, body.Build)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c := &store.Character{AccountID: sess.AccountID, Name: body.Name, Race: arch.Race, Class: arch.Class, Level: 1,
		Stats: stats, HasStats: true, XPScheme: store.XPSchemeRules, Specs: map[string]int{}}
	if st, ok := s.World.Cat.StartFor(arch.Realm, arch.Race, arch.Class); ok {
		c.Zone, c.X, c.Y, c.Z, c.Heading = st.Zone, st.X, st.Y, st.Z, st.Heading
		c.Bind = &store.Place{Zone: st.Zone, X: st.X, Y: st.Y, Z: st.Z, Heading: st.Heading}
	} else {
		c.Zone = s.World.DefaultZone
	}
	c.HP = world.StartHP(c.Class, c.Stats)
	if err := s.Store.CreateCharacter(c); err != nil {
		code := http.StatusBadRequest
		if errors.Is(err, store.ErrNameTaken) {
			code = http.StatusConflict
		}
		http.Error(w, err.Error(), code)
		return
	}
	if err := s.World.GiveStarterKit(c.ID); err != nil {
		s.Log("starter kit for %s: %v", c.Name, err)
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, s.view(c))
}

func (s *Server) character(w http.ResponseWriter, r *http.Request, sess *auth.Session) {
	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/characters/"), 10, 64)
	if err != nil || r.Method != http.MethodDelete {
		http.Error(w, "DELETE /api/characters/{id}", http.StatusBadRequest)
		return
	}
	if err := s.Store.DeleteCharacter(sess.AccountID, id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- the play session ---

const (
	writeWait   = 5 * time.Second
	pingEvery   = 20 * time.Second
	idleTimeout = 120 * time.Second
	maxMsgBytes = 16 << 10
	maxMsgRate  = 120 // messages a second from one client
)

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.Auth.Session(r)
	if !ok {
		http.Error(w, "log in first", http.StatusUnauthorized)
		return
	}
	conn, err := s.up.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetReadLimit(maxMsgBytes)
	fail := func(code, text string) {
		b, _ := json.Marshal(protocol.Err{T: protocol.TErr, Code: code, Text: text})
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		conn.WriteMessage(websocket.TextMessage, b)
	}

	// The first message must be a hello, promptly, with our protocol version.
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		return
	}
	typ, msg, err := protocol.Decode(data)
	hello, isHello := msg.(*protocol.Hello)
	if err != nil || typ != protocol.THello || !isHello {
		fail("hello", "Expected a hello.")
		return
	}
	if hello.V != protocol.Version {
		fail("version", "This page is out of date; reload it.")
		return
	}
	acct, err := s.Store.Account(sess.AccountID)
	if err != nil {
		fail("account", "Unknown account.")
		return
	}
	acct.Admin = sess.Admin
	p, err := s.World.Join(acct, hello.Char)
	if err != nil {
		fail("join", err.Error())
		return
	}
	z := p.Zone
	defer s.World.Leave(p)

	// Writer: the player's outbound queue to the socket, with pings.
	closed := make(chan struct{})
	go func() {
		t := time.NewTicker(pingEvery)
		defer t.Stop()
		for {
			select {
			case <-closed:
				return
			case b := <-p.Out():
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
					conn.Close()
					return
				}
			case <-t.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					conn.Close()
					return
				}
			}
		}
	}()
	defer close(closed)

	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(idleTimeout)); return nil })
	window, count := time.Now(), 0
	for {
		conn.SetReadDeadline(time.Now().Add(idleTimeout))
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if now := time.Now(); now.Sub(window) > time.Second {
			window, count = now, 0
		}
		if count++; count > maxMsgRate {
			fail("rate", "Too many messages.")
			return
		}
		typ, msg, err := protocol.Decode(data)
		if err != nil || typ == "" {
			continue // unknown or malformed: ignored
		}
		s.dispatch(z, p, typ, msg)
	}
}

// dispatch hands a client message to the zone's goroutine.
func (s *Server) dispatch(z *world.Zone, p *world.Player, typ string, msg interface{}) {
	switch m := msg.(type) {
	case *protocol.Input:
		z.Do(func() { p.QueueInput(m) })
	case *protocol.Chat:
		z.Do(func() { s.World.Chat(p, m.Ch, m.Text) })
	case *protocol.Target:
		z.Do(func() { z.SetTarget(p, m.ID) })
	case *protocol.Attack:
		z.Do(func() { z.SetAttack(p, m.On) })
	case *protocol.Style:
		z.Do(func() { z.RequestStyle(p, m.Slot) })
	case *protocol.Loot:
		z.Do(func() { z.TakeLoot(p, m.Corpse, m.Index) })
	case *protocol.Equip:
		z.Do(func() {
			switch {
			case m.To == -1:
				z.DropItem(p, m.From)
			case m.To == -2:
				z.AutoEquip(p, m.From)
			default:
				z.MoveItem(p, m.From, m.To)
			}
		})
	case *protocol.Interact:
		z.Do(func() { z.Interact(p, m.ID) })
	case *protocol.Buy:
		z.Do(func() { z.Buy(p, m.ID) })
	case *protocol.Sell:
		z.Do(func() { z.Sell(p, m.Slot) })
	case *protocol.GM:
		z.Do(func() { s.World.GM(p, m.Line) })
	case *protocol.Train:
		z.Do(func() { z.Train(p, m.Spec, m.N) })
	case *protocol.Restore:
		z.Do(func() { z.Restore(p, m.NPC) })
	}
	if typ == protocol.TPing {
		z.Do(func() { p.Pong() })
	}
	if typ == protocol.TPray {
		z.Do(func() { z.Pray(p) })
	}
	if typ == protocol.TRelease {
		z.Do(func() { z.ReleasePlayer(p) })
	}
}

// StaticDir returns the absolute path of the web directory, for the log.
func (s *Server) StaticDir() string {
	a, err := filepath.Abs(s.WebDir)
	if err != nil {
		return s.WebDir
	}
	return a
}
