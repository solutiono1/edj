package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"edj.cash/edj/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) Config(w http.ResponseWriter, r *http.Request) {
	feeRate := os.Getenv("EDJ_FEE_RATE")
	if feeRate == "" {
		feeRate = "1"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"googleClientId":  os.Getenv("GOOGLE_CLIENT_ID"),
		"receiveAddress":  os.Getenv("EDJ_RECEIVE_ADDRESS"),
		"tokenAddress":    os.Getenv("EDJ_TOKEN_ADDRESS"),
		"derivedAddress":  os.Getenv("EDJ_XPUB") != "",
		"tokenDerivation": os.Getenv("EDJ_TOKEN_XPUB") != "",
		"feeRate":         feeRate,
		"network":         "mainnet",
	})
}

type googleTokenInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Aud           string `json:"aud"`
	EmailVerified string `json:"email_verified"`
}

func (s *Server) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		fail(w, http.StatusServiceUnavailable, "GOOGLE_CLIENT_ID not configured")
		return
	}
	var req struct {
		IDToken string `json:"idToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		fail(w, http.StatusBadRequest, "missing id token")
		return
	}
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(req.IDToken))
	if err != nil {
		fail(w, http.StatusUnauthorized, "token verification failed")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fail(w, http.StatusUnauthorized, "invalid token")
		return
	}
	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil || info.Aud != clientID || info.EmailVerified != "true" || info.Sub == "" {
		fail(w, http.StatusUnauthorized, "token rejected")
		return
	}
	var u *store.User
	var isNew bool
	email := normalizeEmail(info.Email)
	if email != "" {
		if existing, ok := s.Store.UserByEmail(email); ok {
			u, isNew = existing, false
		}
	}
	if u == nil {
		u, isNew = s.Store.UpsertUser(&store.User{Sub: info.Sub, Email: email, Name: info.Name})
	}
	s.newSession(w, u, isNew)
}

const (
	minPasswordLength = 10
	maxPasswordLength = 72 // bcrypt only uses the first 72 bytes
)

var (
	emailAuthLimiter = &attemptLimiter{limit: 15, window: 15 * time.Minute}
	dummyHash, _     = bcrypt.GenerateFromPassword([]byte("pagor-invalid-credential"), bcrypt.DefaultCost)
)

type attemptLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

func (l *attemptLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.hits == nil {
		l.hits = make(map[string][]time.Time)
	}
	cutoff := time.Now().Add(-l.window)
	recent := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= l.limit {
		l.hits[key] = recent
		return false
	}
	l.hits[key] = append(recent, time.Now())
	return true
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func validEmail(email string) bool {
	if email == "" || len(email) > 254 || strings.ContainsAny(email, " \t\r\n") {
		return false
	}
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func passwordError(password string) string {
	if len(password) < minPasswordLength {
		return "Use at least 10 characters for your password."
	}
	if len(password) > maxPasswordLength {
		return "Use at most 72 characters for your password."
	}
	return ""
}

func nameFromEmail(email string) string {
	local := email
	if i := strings.IndexByte(email, '@'); i > 0 {
		local = email[:i]
	}
	parts := strings.Fields(strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(local))
	if len(parts) == 0 {
		return "Member"
	}
	var b strings.Builder
	for _, part := range parts {
		runes := []rune(part)
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strings.ToUpper(string(runes[0])) + strings.ToLower(string(runes[1:])))
	}
	return b.String()
}

func (s *Server) newSession(w http.ResponseWriter, user *store.User, isNew bool) {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	s.Store.PutSession(token, user.Sub)
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user, "isNew": isNew})
}

type emailAuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (s *Server) EmailSignup(w http.ResponseWriter, r *http.Request) {
	if !emailAuthLimiter.allow(clientIP(r)) {
		fail(w, http.StatusTooManyRequests, "Too many attempts. Please try again in a few minutes.")
		return
	}
	var req emailAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "Invalid request.")
		return
	}
	email := normalizeEmail(req.Email)
	if !validEmail(email) {
		fail(w, http.StatusBadRequest, "Enter a valid email address.")
		return
	}
	if msg := passwordError(req.Password); msg != "" {
		fail(w, http.StatusBadRequest, msg)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = nameFromEmail(email)
	}
	if len(name) > 80 {
		name = name[:80]
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Could not create your account. Try again.")
		return
	}
	user, created := s.Store.CreateEmailUser(&store.User{
		Sub: "email:" + email, Email: email, Name: name,
	}, string(hash))
	if !created {
		fail(w, http.StatusConflict, "An account with that email already exists. Log in instead.")
		return
	}
	s.newSession(w, user, true)
}

func (s *Server) EmailLogin(w http.ResponseWriter, r *http.Request) {
	if !emailAuthLimiter.allow(clientIP(r)) {
		fail(w, http.StatusTooManyRequests, "Too many attempts. Please try again in a few minutes.")
		return
	}
	var req emailAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "Invalid request.")
		return
	}
	email := normalizeEmail(req.Email)
	if !validEmail(email) {
		fail(w, http.StatusBadRequest, "Enter a valid email address.")
		return
	}
	user, ok := s.Store.UserByEmail(email)
	if !ok {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
		fail(w, http.StatusUnauthorized, "Invalid email or password.")
		return
	}
	hash, ok := s.Store.PasswordHash(user.Sub)
	if !ok || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		fail(w, http.StatusUnauthorized, "Invalid email or password.")
		return
	}
	s.newSession(w, user, false)
}

func (s *Server) userFromRequest(r *http.Request) (*store.User, string, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, "", false
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	sub, ok := s.Store.Session(token)
	if !ok {
		return nil, "", false
	}
	u, ok := s.Store.User(sub)
	return u, token, ok
}

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	u, _, ok := s.userFromRequest(r)
	if !ok {
		fail(w, http.StatusUnauthorized, "not signed in")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	_, token, ok := s.userFromRequest(r)
	if ok {
		s.Store.DeleteSession(token)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (s *Server) Notifications(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Store.Notifications())
}

// Assistant is a rules-based guide over real app data and real routes.
// No LLM and no fabricated answers.
func (s *Server) Assistant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "invalid message")
		return
	}
	m := strings.ToLower(req.Message)
	answer := "I can help with private links, payroll, proofs, and your wallet. Try asking 'how do I create a link?'"
	link := "/#how"
	switch {
	case strings.Contains(m, "link"):
		answer = "Open Links in the sidebar, enter an amount and asset, then create the link. The amount is not published on-chain."
		link = "/links"
	case strings.Contains(m, "payroll") || strings.Contains(m, "pay people") || strings.Contains(m, "team"):
		answer = "Upload a CSV with columns name,address,amount. Review the public view before paying."
		link = "/payroll"
	case strings.Contains(m, "proof") || strings.Contains(m, "prove"):
		answer = "Create a signed statement with a subject, a threshold and an expiry. It is not a zero-knowledge proof."
		link = "/proofs"
	case strings.Contains(m, "wallet") || strings.Contains(m, "address"):
		answer = "Your receiving address is in Wallet settings. Never send funds to an address you have not verified."
		link = "/wallet"
	case strings.Contains(m, "anon") || strings.Contains(m, "private"):
		answer = "pagor reduces what is published on-chain. It is not anonymity, and we never claim it is."
	}
	writeJSON(w, http.StatusOK, map[string]string{"answer": answer, "link": link})
}

// RequireUser rejects requests without a valid pagor session token.
func (s *Server) RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := s.userFromRequest(r); !ok {
			fail(w, http.StatusUnauthorized, "not signed in")
			return
		}
		next(w, r)
	}
}
