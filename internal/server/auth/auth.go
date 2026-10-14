// Package auth logs players in: a Discord OAuth2 code flow, or an
// administrator password, both ending in the same signed session cookie.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"daocweb/internal/server/store"
)

// CookieName is the session cookie.
const CookieName = "daoc_session"

// Config holds what the environment supplies. Nothing here has a default
// that lets anyone in: with no admin password and no Discord application the
// server refuses to start.
type Config struct {
	AdminHash     string // bcrypt hash of the admin password
	AdminPassword string // plain password, hashed at start; AdminHash wins
	DiscordID     string
	DiscordSecret string
	DiscordBase   string // https://discord.com by default; a test stub in tests
	PublicURL     string // where the browser reaches us, e.g. http://localhost:8777
	SessionKey    []byte
	SessionTTL    time.Duration
	Secure        bool // set the Secure flag on cookies (when served over HTTPS)
	Log           func(format string, a ...any)
}

// ErrNotConfigured is returned by New when no login method is set up.
var ErrNotConfigured = errors.New("no login is configured: set DAOC_ADMIN_PASSWORD (or DAOC_ADMIN_HASH) for local play, " +
	"and/or DAOC_DISCORD_CLIENT_ID, DAOC_DISCORD_CLIENT_SECRET and DAOC_PUBLIC_URL for Discord")

// Session is what a valid cookie says.
type Session struct {
	AccountID int64
	Admin     bool
	Expires   time.Time
}

// Auth is the login service.
type Auth struct {
	cfg   Config
	store *store.Store
	hash  []byte
	rl    *limiter
	http  *http.Client
	now   func() time.Time
}

// New checks the configuration and builds the service.
func New(cfg Config, st *store.Store) (*Auth, error) {
	if cfg.Log == nil {
		cfg.Log = func(string, ...any) {}
	}
	a := &Auth{cfg: cfg, store: st, rl: newLimiter(5, 5*time.Minute), http: &http.Client{Timeout: 10 * time.Second}, now: time.Now}
	switch {
	case cfg.AdminHash != "":
		a.hash = []byte(cfg.AdminHash)
	case cfg.AdminPassword != "":
		h, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		a.hash = h
	}
	if a.hash == nil && (cfg.DiscordID == "" || cfg.DiscordSecret == "") {
		return nil, ErrNotConfigured
	}
	if len(cfg.SessionKey) < 16 {
		return nil, errors.New("auth: session key must be at least 16 bytes")
	}
	if a.cfg.SessionTTL == 0 {
		a.cfg.SessionTTL = 24 * time.Hour * 7
	}
	if a.cfg.DiscordBase == "" {
		a.cfg.DiscordBase = "https://discord.com"
	}
	return a, nil
}

// HasAdmin and HasDiscord say which login methods are on.
func (a *Auth) HasAdmin() bool   { return a.hash != nil }
func (a *Auth) HasDiscord() bool { return a.cfg.DiscordID != "" && a.cfg.DiscordSecret != "" }

// --- sessions ---

func (a *Auth) sign(payload string) string {
	m := hmac.New(sha256.New, a.cfg.SessionKey)
	m.Write([]byte(payload))
	return hex.EncodeToString(m.Sum(nil))
}

// issue writes the session cookie.
func (a *Auth) issue(w http.ResponseWriter, accountID int64, admin bool) {
	exp := a.now().Add(a.cfg.SessionTTL)
	adm := "0"
	if admin {
		adm = "1"
	}
	payload := fmt.Sprintf("%d|%s|%d", accountID, adm, exp.Unix())
	val := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + a.sign(payload)
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: val, Path: "/", Expires: exp,
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: a.cfg.Secure})
}

// Logout clears the cookie.
func (a *Auth) Logout(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
}

// Session reads and verifies a request's cookie.
func (a *Auth) Session(r *http.Request) (*Session, bool) {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return nil, false
	}
	i := strings.IndexByte(c.Value, '.')
	if i < 0 {
		return nil, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(c.Value[:i])
	if err != nil {
		return nil, false
	}
	payload := string(raw)
	if subtle.ConstantTimeCompare([]byte(a.sign(payload)), []byte(c.Value[i+1:])) != 1 {
		return nil, false
	}
	p := strings.Split(payload, "|")
	if len(p) != 3 {
		return nil, false
	}
	id, e1 := strconv.ParseInt(p[0], 10, 64)
	exp, e2 := strconv.ParseInt(p[2], 10, 64)
	if e1 != nil || e2 != nil || a.now().Unix() > exp {
		return nil, false
	}
	return &Session{AccountID: id, Admin: p[1] == "1", Expires: time.Unix(exp, 0)}, true
}

// Require wraps a handler so only logged-in requests reach it.
func (a *Auth) Require(next func(http.ResponseWriter, *http.Request, *Session)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.Session(r)
		if !ok {
			http.Error(w, "log in first", http.StatusUnauthorized)
			return
		}
		next(w, r, s)
	}
}

// --- admin password ---

// AdminLogin is POST /auth/admin with {"password": "..."}.
func (a *Auth) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if !a.HasAdmin() {
		http.Error(w, "admin login is off", http.StatusNotFound)
		return
	}
	ip := clientIP(r)
	if !a.rl.allow(ip, a.now()) {
		http.Error(w, "too many attempts; wait a few minutes", http.StatusTooManyRequests)
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if bcrypt.CompareHashAndPassword(a.hash, []byte(body.Password)) != nil {
		a.rl.fail(ip, a.now())
		http.Error(w, "wrong password", http.StatusUnauthorized)
		return
	}
	a.rl.reset(ip)
	acct, err := a.store.AccountFor("admin:local", "Administrator", true)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	a.issue(w, acct.ID, true)
	w.WriteHeader(http.StatusNoContent)
}

func clientIP(r *http.Request) string {
	h, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return h
}

// limiter allows `max` failures per address within `window`.
type limiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	fails  map[string][]time.Time
}

func newLimiter(max int, window time.Duration) *limiter {
	return &limiter{max: max, window: window, fails: map[string][]time.Time{}}
}

func (l *limiter) prune(ip string, now time.Time) []time.Time {
	var keep []time.Time
	for _, t := range l.fails[ip] {
		if now.Sub(t) < l.window {
			keep = append(keep, t)
		}
	}
	l.fails[ip] = keep
	return keep
}

func (l *limiter) allow(ip string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.prune(ip, now)) < l.max
}

func (l *limiter) fail(ip string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fails[ip] = append(l.prune(ip, now), now)
}

func (l *limiter) reset(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, ip)
}

// --- Discord ---

const stateCookie = "daoc_oauth_state"

func (a *Auth) redirectURI() string {
	return strings.TrimSuffix(a.cfg.PublicURL, "/") + "/auth/discord/callback"
}

// DiscordStart is GET /auth/discord: it sends the browser to Discord.
func (a *Auth) DiscordStart(w http.ResponseWriter, r *http.Request) {
	if !a.HasDiscord() {
		http.Error(w, "Discord login is off", http.StatusNotFound)
		return
	}
	b := make([]byte, 16)
	rand.Read(b)
	state := hex.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{Name: stateCookie, Value: state, Path: "/", MaxAge: 600,
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: a.cfg.Secure})
	q := url.Values{
		"client_id": {a.cfg.DiscordID}, "redirect_uri": {a.redirectURI()},
		"response_type": {"code"}, "scope": {"identify"}, "state": {state},
	}
	a.cfg.Log("discord login started; redirect_uri=%s", a.redirectURI())
	http.Redirect(w, r, a.cfg.DiscordBase+"/oauth2/authorize?"+q.Encode(), http.StatusFound)
}

// DiscordCallback is GET /auth/discord/callback?code=&state=.
func (a *Auth) DiscordCallback(w http.ResponseWriter, r *http.Request) {
	if !a.HasDiscord() {
		http.Error(w, "Discord login is off", http.StatusNotFound)
		return
	}
	c, err := r.Cookie(stateCookie)
	state := r.URL.Query().Get("state")
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(c.Value), []byte(state)) != 1 {
		a.cfg.Log("discord callback: state cookie missing or wrong (cookie %v, state given %v); use the same host name throughout, e.g. localhost", err == nil, state != "")
		http.Error(w, "login expired; try again", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		a.cfg.Log("discord callback: no code; Discord said error=%q description=%q", r.URL.Query().Get("error"), r.URL.Query().Get("error_description"))
		http.Error(w, "Discord did not approve the login", http.StatusBadRequest)
		return
	}
	id, name, err := a.exchange(code)
	if err != nil {
		a.cfg.Log("discord callback: %v", err)
		http.Error(w, "Discord login failed", http.StatusBadGateway)
		return
	}
	acct, err := a.store.AccountFor("discord:"+id, name, false)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: stateCookie, Value: "", Path: "/", MaxAge: -1})
	a.cfg.Log("discord login: %s (account %d)", name, acct.ID)
	a.issue(w, acct.ID, false)
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

// exchange trades the code for a token and asks who it belongs to.
func (a *Auth) exchange(code string) (id, name string, err error) {
	form := url.Values{
		"client_id": {a.cfg.DiscordID}, "client_secret": {a.cfg.DiscordSecret},
		"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {a.redirectURI()},
	}
	resp, err := a.http.PostForm(a.cfg.DiscordBase+"/api/oauth2/token", form)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if resp.StatusCode != 200 || json.NewDecoder(resp.Body).Decode(&tok) != nil || tok.AccessToken == "" {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 400))
		return "", "", fmt.Errorf("token exchange: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}
	req, _ := http.NewRequest("GET", a.cfg.DiscordBase+"/api/users/@me", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r2, err := a.http.Do(req)
	if err != nil {
		return "", "", err
	}
	defer r2.Body.Close()
	var u struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
	}
	if r2.StatusCode != 200 || json.NewDecoder(r2.Body).Decode(&u) != nil || u.ID == "" {
		return "", "", fmt.Errorf("user lookup: %s", r2.Status)
	}
	name = u.GlobalName
	if name == "" {
		name = u.Username
	}
	return u.ID, name, nil
}
