package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"edj.cash/edj/internal/chain"
	"edj.cash/edj/internal/store"
)

const testXpub = "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8"

func createLink(t *testing.T, s *Server) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(`{"amount":25000,"note":"Invoice 42"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.CreateLink(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create link status = %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestCreateLinkDerivesFreshAddress(t *testing.T) {
	t.Setenv("EDJ_XPUB", testXpub)
	t.Setenv("EDJ_TOKEN_XPUB", "")
	s := &Server{Store: store.New(), Chain: chain.BCH{ReceiveAddress: "bitcoincash:fallback"}}

	first := createLink(t, s)
	second := createLink(t, s)

	addr1, _ := first["address"].(string)
	addr2, _ := second["address"].(string)
	if !strings.HasPrefix(addr1, "bitcoincash:") || !strings.HasPrefix(addr2, "bitcoincash:") {
		t.Fatalf("addresses not mainnet cashaddr: %s %s", addr1, addr2)
	}
	if addr1 == addr2 {
		t.Fatal("two links reused the same address")
	}
	idx1, ok1 := first["derivationIndex"]
	idx2, ok2 := second["derivationIndex"]
	if !ok1 || !ok2 {
		t.Fatal("derivation index missing from create response")
	}
	if idx1 == idx2 {
		t.Fatal("two links reused the same derivation index")
	}
}

func TestCreateLinkFallsBackToConfiguredAddress(t *testing.T) {
	t.Setenv("EDJ_XPUB", "")
	s := &Server{Store: store.New(), Chain: chain.BCH{ReceiveAddress: "bitcoincash:configured"}}
	body := createLink(t, s)
	if body["address"] != "bitcoincash:configured" {
		t.Fatalf("fallback address = %v", body["address"])
	}
	if body["derivationIndex"] != float64(0) {
		t.Fatalf("fallback derivation index = %v, want 0", body["derivationIndex"])
	}
}

func TestFreshAddressRejectsInvalidXpub(t *testing.T) {
	t.Setenv("EDJ_XPUB", "garbage")
	s := &Server{Store: store.New(), Chain: chain.BCH{ReceiveAddress: "bitcoincash:configured"}}
	if _, _, err := s.freshAddress(context.Background(), nil); err == nil {
		t.Fatal("invalid xpub was accepted")
	}
}
