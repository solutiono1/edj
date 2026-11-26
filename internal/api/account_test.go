package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"edj.cash/edj/internal/core"
	"edj.cash/edj/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const accountPassword = "a long password"

func accountServer() (*Server, *store.Memory, string) {
	s := &Server{Store: store.New()}
	m := s.Store.(*store.Memory)
	hash, _ := bcrypt.GenerateFromPassword([]byte(accountPassword), bcrypt.MinCost)
	u := &store.User{Sub: "email:ada@example.com", Email: "ada@example.com", Name: "Ada", Status: "active"}
	m.CreateEmailUser(u, string(hash))
	m.PutSession("session-abc", u.Sub)
	return s, m, "session-abc"
}

func accountPost(s *Server, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	switch path {
	case "/api/account/disable":
		s.DisableAccount(rec, req)
	case "/api/account/delete":
		s.ScheduleDelete(rec, req)
	case "/api/account/restore":
		s.RestoreAccount(rec, req)
	}
	return rec
}

func loginAgain(t *testing.T, s *Server) string {
	t.Helper()
	body := `{"email":"ada@example.com","password":"` + accountPassword + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.EmailLogin(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d: %s", rec.Code, rec.Body.String())
	}
	var data map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	token, _ := data["token"].(string)
	return token
}

func TestDisableRevokesSession(t *testing.T) {
	s, _, token := accountServer()
	if rec := accountPost(s, "/api/account/disable", token); rec.Code != http.StatusOK {
		t.Fatalf("disable status = %d", rec.Code)
	}
	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	meRec := httptest.NewRecorder()
	s.Me(meRec, meReq)
	if meRec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled account still has a working session: %d", meRec.Code)
	}
}

func TestDeleteThenRestoreWithinWindow(t *testing.T) {
	s, _, token := accountServer()
	rec := accountPost(s, "/api/account/delete", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "deleted" || body["deletedAt"] == nil {
		t.Fatalf("bad delete response: %v", body)
	}

	newToken := loginAgain(t, s)
	restore := accountPost(s, "/api/account/restore", newToken)
	if restore.Code != http.StatusOK {
		t.Fatalf("restore status = %d: %s", restore.Code, restore.Body.String())
	}
	var restored map[string]any
	_ = json.Unmarshal(restore.Body.Bytes(), &restored)
	restoredToken, _ := restored["token"].(string)
	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+restoredToken)
	meRec := httptest.NewRecorder()
	s.Me(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("restored account cannot use session: %d", meRec.Code)
	}
}

func TestRestoreAfterWindowIsRejected(t *testing.T) {
	s, m, token := accountServer()
	accountPost(s, "/api/account/delete", token)
	u, _ := m.User("email:ada@example.com")
	past := time.Now().UTC().Add(-31 * 24 * time.Hour)
	u.DeletedAt = &past
	m.UpdateUser(u)
	rec := accountPost(s, "/api/account/restore", loginAgain(t, s))
	if rec.Code != http.StatusGone {
		t.Fatalf("restore after window status = %d, want 410", rec.Code)
	}
}

func TestPurgeRemovesUserSessionsAndRecords(t *testing.T) {
	_, m, token := accountServer()
	u, _ := m.User("email:ada@example.com")
	past := time.Now().UTC().Add(-31 * 24 * time.Hour)
	u.Status = "deleted"
	u.DeletedAt = &past
	m.UpdateUser(u)
	m.PutLink(&core.Link{ID: "L_test", Address: "bitcoincash:qtest", Amount: 100, OwnerSub: u.Sub})
	m.PutRun(&core.PayrollRun{ID: "PR_test", OwnerSub: u.Sub})
	m.PutProof(&core.DisclosureProof{ID: "P_test", OwnerSub: u.Sub})

	n := m.PurgeAccountsBefore(time.Now().UTC().Add(-30 * 24 * time.Hour))
	if n != 1 {
		t.Fatalf("purged %d accounts, want 1", n)
	}
	if _, ok := m.User(u.Sub); ok {
		t.Fatal("user still exists")
	}
	if _, ok := m.Session(token); ok {
		t.Fatal("session still exists")
	}
	if _, ok := m.GetLink("L_test"); ok {
		t.Fatal("link still exists")
	}
	if _, ok := m.GetRun("PR_test"); ok {
		t.Fatal("run still exists")
	}
	if _, ok := m.GetProof("P_test"); ok {
		t.Fatal("proof still exists")
	}
}
