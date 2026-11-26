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

func TestPrivatePayrollStoresOnlyOpaqueBlob(t *testing.T) {
	s := &Server{Store: store.New()}
	_ = s.Store.(*store.Memory)
	u := &store.User{Sub: "email:owner@example.com", Email: "owner@example.com", Name: "Owner", Status: "active"}
	s.Store.UpsertUser(u)
	s.Store.PutSession("owner-token", u.Sub)

	body := `{"salt":"c2FsdA==","iv":"aXY=","ciphertext":"ZW5jcnlwdGVk","rowCount":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/private-payroll", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer owner-token")
	rec := httptest.NewRecorder()
	s.UploadPrivatePayroll(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload status = %d: %s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	id, _ := created["id"].(string)
	if !strings.HasPrefix(id, "PP_") {
		t.Fatalf("bad id: %s", id)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/private-payroll/"+id, nil)
	getReq.SetPathValue("id", id)
	getReq.Header.Set("Authorization", "Bearer owner-token")
	getRec := httptest.NewRecorder()
	s.GetPrivatePayroll(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d", getRec.Code)
	}
	var stored map[string]any
	_ = json.Unmarshal(getRec.Body.Bytes(), &stored)
	if stored["ciphertext"] != "ZW5jcnlwdGVk" {
		t.Fatalf("ciphertext not stored: %v", stored)
	}
	if _, leaked := stored["ownerSub"]; leaked {
		t.Fatal("owner leaked in response")
	}
	for _, key := range []string{"Ada", "bitcoincash", "125000"} {
		if strings.Contains(getRec.Body.String(), key) {
			t.Fatalf("plaintext leaked: %s", key)
		}
	}
}

func TestPrivatePayrollRejectsOtherOwner(t *testing.T) {
	s := &Server{Store: store.New()}
	a := &store.User{Sub: "email:a@example.com", Email: "a@example.com", Name: "A", Status: "active"}
	b := &store.User{Sub: "email:b@example.com", Email: "b@example.com", Name: "B", Status: "active"}
	s.Store.UpsertUser(a)
	s.Store.UpsertUser(b)
	s.Store.PutSession("a-token", a.Sub)
	s.Store.PutSession("b-token", b.Sub)

	body := `{"salt":"c2FsdA==","iv":"aXY=","ciphertext":"ZW5jcnlwdGVk","rowCount":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/private-payroll", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer a-token")
	rec := httptest.NewRecorder()
	s.UploadPrivatePayroll(rec, req)
	var created map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	id, _ := created["id"].(string)

	getReq := httptest.NewRequest(http.MethodGet, "/api/private-payroll/"+id, nil)
	getReq.SetPathValue("id", id)
	getReq.Header.Set("Authorization", "Bearer b-token")
	getRec := httptest.NewRecorder()
	s.GetPrivatePayroll(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("other owner status = %d, want 404", getRec.Code)
	}
}
