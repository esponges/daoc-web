package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"daocweb/internal/server/store"
)

var key = []byte("0123456789abcdef0123456789abcdef")

func newAuth(t *testing.T, cfg Config) (*Auth, *store.Store) {
	t.Helper()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	cfg.SessionKey = key
	a, err := New(cfg, st)
	if err != nil {
		t.Fatal(err)
	}
	return a, st
}

func login(a *Auth, pw, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/auth/admin", strings.NewReader(`{"password":"`+pw+`"}`))
	req.RemoteAddr = ip + ":1234"
	w := httptest.NewRecorder()
	a.AdminLogin(w, req)
	return w
}

func TestNotConfigured(t *testing.T) {
	st, _ := store.Open(":memory:")
	defer st.Close()
	if _, err := New(Config{SessionKey: key}, st); err != ErrNotConfigured {
		t.Errorf("err = %v, want ErrNotConfigured", err)
	}
}

func TestAdminLoginAndSession(t *testing.T) {
	a, _ := newAuth(t, Config{AdminPassword: "hunter2-hunter2"})
	w := login(a, "hunter2-hunter2", "10.0.0.1")
	if w.Code != 204 {
		t.Fatalf("right password: %d", w.Code)
	}
	req := httptest.NewRequest("GET", "/", nil)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	s, ok := a.Session(req)
	if !ok || !s.Admin || s.AccountID == 0 {
		t.Errorf("session = %+v, %v", s, ok)
	}
	// a tampered cookie, a stranger's cookie and a missing one all fail
	bad := httptest.NewRequest("GET", "/", nil)
	bad.AddCookie(&http.Cookie{Name: CookieName, Value: strings.Replace(w.Result().Cookies()[0].Value, "A", "B", 1) + "x"})
	if _, ok := a.Session(bad); ok {
		t.Error("tampered cookie accepted")
	}
	if _, ok := a.Session(httptest.NewRequest("GET", "/", nil)); ok {
		t.Error("no cookie accepted")
	}
	// expiry
	a.now = func() time.Time { return time.Now().Add(30 * 24 * time.Hour) }
	if _, ok := a.Session(req); ok {
		t.Error("expired session accepted")
	}
}

func TestWrongPasswordRateLimited(t *testing.T) {
	a, _ := newAuth(t, Config{AdminPassword: "correct horse"})
	for i := 0; i < 5; i++ {
		if w := login(a, "nope", "10.0.0.2"); w.Code != 401 {
			t.Fatalf("attempt %d: %d", i, w.Code)
		}
	}
	// even the right password is refused once the address is limited
	if w := login(a, "correct horse", "10.0.0.2"); w.Code != 429 {
		t.Errorf("after 5 failures: %d, want 429", w.Code)
	}
	// another address is unaffected
	if w := login(a, "correct horse", "10.0.0.3"); w.Code != 204 {
		t.Errorf("other address: %d", w.Code)
	}
}

func TestRequireWithoutCookie(t *testing.T) {
	a, _ := newAuth(t, Config{AdminPassword: "pw"})
	h := a.Require(func(http.ResponseWriter, *http.Request, *Session) { t.Error("handler must not run") })
	w := httptest.NewRecorder()
	h(w, httptest.NewRequest("GET", "/api/me", nil))
	if w.Code != 401 {
		t.Errorf("code = %d", w.Code)
	}
}

func TestDiscordFlow(t *testing.T) {
	// A stub Discord: token endpoint and user endpoint.
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/oauth2/token":
			r.ParseForm()
			if r.Form.Get("code") != "good" || r.Form.Get("client_secret") != "shh" {
				http.Error(w, "bad", 400)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"access_token": "tok"})
		case "/api/users/@me":
			if r.Header.Get("Authorization") != "Bearer tok" {
				http.Error(w, "no", 401)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "4242", "username": "grimm", "global_name": "Grimm"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer stub.Close()
	a, st := newAuth(t, Config{DiscordID: "cid", DiscordSecret: "shh", DiscordBase: stub.URL, PublicURL: "http://localhost:8777"})

	// start: redirect carries our client id, redirect uri and a state cookie
	w := httptest.NewRecorder()
	a.DiscordStart(w, httptest.NewRequest("GET", "/auth/discord", nil))
	loc, _ := url.Parse(w.Header().Get("Location"))
	if w.Code != 302 || loc.Query().Get("client_id") != "cid" ||
		loc.Query().Get("redirect_uri") != "http://localhost:8777/auth/discord/callback" {
		t.Fatalf("start: %d %s", w.Code, w.Header().Get("Location"))
	}
	state := loc.Query().Get("state")
	sc := w.Result().Cookies()[0]

	cb := func(code, st string, cookie *http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/auth/discord/callback?code="+code+"&state="+st, nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		a.DiscordCallback(w, req)
		return w
	}
	// a state that does not match is refused (CSRF)
	if w := cb("good", "forged", sc); w.Code != 400 {
		t.Errorf("forged state: %d", w.Code)
	}
	if w := cb("good", state, nil); w.Code != 400 {
		t.Errorf("no state cookie: %d", w.Code)
	}
	if w := cb("rejected", state, sc); w.Code != 502 {
		t.Errorf("bad code: %d", w.Code)
	}
	w = cb("good", state, sc)
	if w.Code != 302 {
		t.Fatalf("callback: %d", w.Code)
	}
	var sess *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == CookieName {
			sess = c
		}
	}
	if sess == nil {
		t.Fatal("no session cookie")
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(sess)
	s, ok := a.Session(req)
	if !ok || s.Admin {
		t.Fatalf("session = %+v %v", s, ok)
	}
	acct, _ := st.Account(s.AccountID)
	if acct.DiscordID != "discord:4242" || acct.Name != "Grimm" {
		t.Errorf("account = %+v", acct)
	}
}
