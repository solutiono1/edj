package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	"edj.cash/edj/internal/store"
)

func (s *Server) Config(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"googleClientId": os.Getenv("GOOGLE_CLIENT_ID"),
		"receiveAddress": os.Getenv("EDJ_RECEIVE_ADDRESS"),
		"tokenAddress":   os.Getenv("EDJ_TOKEN_ADDRESS"),
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
	u, isNew := s.Store.UpsertUser(&store.User{Sub: info.Sub, Email: info.Email, Name: info.Name})
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	s.Store.PutSession(token, u.Sub)
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": u, "isNew": isNew})
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
