package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"edj.cash/edj/internal/core"
	"edj.cash/edj/internal/store"
)

func TestPayrollResultAndExport(t *testing.T) {
	s := &Server{Store: store.New()}
	run := core.NewPayrollRun([]core.PayrollRow{
		{Name: "Ada", Address: "bitcoincash:qaaa", Amount: 1000},
	})
	s.Store.PutRun(run)
	path := "/api/payroll/" + run.ID

	post := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path+"/result", bytes.NewBufferString(body))
		req.SetPathValue("id", run.ID)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		s.PayrollResult(rec, req)
		return rec
	}

	first := post(`{"index":0,"status":"broadcast","txid":"abc123"}`)
	if first.Code != http.StatusOK {
		t.Fatalf("result status = %d: %s", first.Code, first.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(first.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["done"] != true || payload["status"] != "complete" {
		t.Fatalf("run not marked complete: %v", payload)
	}

	duplicate := post(`{"index":0,"status":"broadcast","txid":"different"}`)
	if duplicate.Code != http.StatusBadRequest {
		t.Fatalf("duplicate result status = %d, want 400", duplicate.Code)
	}

	req := httptest.NewRequest(http.MethodGet, path+"/export", nil)
	req.SetPathValue("id", run.ID)
	rec := httptest.NewRecorder()
	s.PayrollExport(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("export status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "abc123") {
		t.Fatalf("export missing txid: %s", rec.Body.String())
	}
	if !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("bad content type: %s", rec.Header().Get("Content-Type"))
	}
}
