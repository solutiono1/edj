package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"edj.cash/edj/internal/store"
)

func authServer() *Server {
	return &Server{Store: store.New()}
}

func postJSON(t *testing.T, s *Server, path, ip, body string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", ip)
	rec := httptest.NewRecorder()

	switch path {
	case "/api/auth/signup":
		s.EmailSignup(rec, req)
	case "/api/auth/login":
		s.EmailLogin(rec, req)
	default:
		t.Fatalf("unexpected path %s", path)
	}

	var decoded map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return rec, decoded
}

func getWithToken(t *testing.T, s *Server, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	s.Me(rec, req)
	return rec
}

func TestEmailSignupAndLogin(t *testing.T) {
	s := authServer()

	rec, body := postJSON(t, s, "/api/auth/signup", "10.0.0.1",
		`{"name":"Ada Lovelace","email":"ADA@Example.COM","password":"correct horse battery"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("signup status = %d, body = %v", rec.Code, body)
	}
	token, ok := body["token"].(string)
	if !ok || token == "" {
		t.Fatalf("no token in signup response: %v", body)
	}
	user, ok := body["user"].(map[string]any)
	if !ok {
		t.Fatalf("no user in signup response: %v", body)
	}
	if user["email"] != "ada@example.com" {
		t.Fatalf("email not normalized: %v", user["email"])
	}
	if _, leaked := user["password_hash"]; leaked {
		t.Fatal("password hash leaked to the browser")
	}
	if raw := rec.Body.String(); strings.Contains(raw, "correct horse battery") || strings.Contains(raw, "$2a$") {
		t.Fatalf("credentials leaked in response: %s", raw)
	}

	me := getWithToken(t, s, "/api/me", token)
	if me.Code != http.StatusOK {
		t.Fatalf("me with session token status = %d", me.Code)
	}

	wrong, _ := postJSON(t, s, "/api/auth/login", "10.0.0.2",
		`{"email":"ada@example.com","password":"wrong password"}`)
	if wrong.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password status = %d, want 401", wrong.Code)
	}

	right, _ := postJSON(t, s, "/api/auth/login", "10.0.0.3",
		`{"email":"ADA@example.com","password":"correct horse battery"}`)
	if right.Code != http.StatusOK {
		t.Fatalf("correct password status = %d, want 200", right.Code)
	}
}

func TestEmailSignupRejectsDuplicatesAndWeakPasswords(t *testing.T) {
	s := authServer()
	signup := `{"email":"grace@example.com","password":"a long password"}`

	first, _ := postJSON(t, s, "/api/auth/signup", "10.0.1.1", signup)
	if first.Code != http.StatusOK {
		t.Fatalf("first signup status = %d", first.Code)
	}
	dup, _ := postJSON(t, s, "/api/auth/signup", "10.0.1.2", signup)
	if dup.Code != http.StatusConflict {
		t.Fatalf("duplicate signup status = %d, want 409", dup.Code)
	}

	weak, _ := postJSON(t, s, "/api/auth/signup", "10.0.1.3",
		`{"email":"short@example.com","password":"short"}`)
	if weak.Code != http.StatusBadRequest {
		t.Fatalf("weak password status = %d, want 400", weak.Code)
	}
	badEmail, _ := postJSON(t, s, "/api/auth/signup", "10.0.1.4",
		`{"email":"not-an-email","password":"a long password"}`)
	if badEmail.Code != http.StatusBadRequest {
		t.Fatalf("bad email status = %d, want 400", badEmail.Code)
	}
}

func TestEmailLoginDoesNotRevealMissingAccount(t *testing.T) {
	s := authServer()
	missing, body := postJSON(t, s, "/api/auth/login", "10.0.2.1",
		`{"email":"nobody@example.com","password":"some password"}`)
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("missing account status = %d, want 401", missing.Code)
	}
	errMsg, _ := body["error"].(map[string]any)["message"].(string)
	if strings.Contains(strings.ToLower(errMsg), "no account") || strings.Contains(strings.ToLower(errMsg), "not found") {
		t.Fatalf("login reveals account existence: %s", errMsg)
	}
}
