package api

import (
	"encoding/json"
	"net/http"

	"edj.cash/edj/internal/chain"
	"edj.cash/edj/internal/core"
	"edj.cash/edj/internal/store"
)

type Server struct {
	Chain  chain.Chain
	Store  store.Store
	Signer *core.Signer
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"error": map[string]string{"code": "ERROR", "message": msg}})
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "engine": "edj"})
}

type linkRequest struct {
	Amount int64   `json:"amount"`
	Token  *string `json:"token"`
	Note   string  `json:"note"`
}

func (s *Server) CreateLink(w http.ResponseWriter, r *http.Request) {
	var req linkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		fail(w, http.StatusBadRequest, "invalid amount")
		return
	}
	addr, err := s.Chain.AddressFor(r.Context(), req.Token)
	if err != nil {
		fail(w, http.StatusInternalServerError, "address generation failed")
		return
	}
	l := core.NewLink(addr, req.Amount, req.Token, req.Note)
	s.Store.PutLink(l)
	writeJSON(w, http.StatusCreated, map[string]any{"id": l.ID, "claimUrl": "/l/" + l.ID, "address": l.Address, "status": l.Status})
}

func (s *Server) GetLink(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	l, ok := s.Store.GetLink(id)
	if !ok {
		fail(w, http.StatusNotFound, "link not found")
		return
	}
	writeJSON(w, http.StatusOK, l)
}

func (s *Server) UploadPayroll(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		fail(w, http.StatusBadRequest, "invalid upload")
		return
	}
	f, _, err := r.FormFile("csv")
	if err != nil {
		fail(w, http.StatusBadRequest, "missing csv")
		return
	}
	defer f.Close()
	rows, err := core.ParsePayroll(f)
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	run := core.NewPayrollRun(rows)
	s.Store.PutRun(run)
	writeJSON(w, http.StatusCreated, map[string]any{"id": run.ID, "rows": len(rows), "total": run.Total, "reviewUrl": "/payroll/" + run.ID})
}

func (s *Server) GetPayroll(w http.ResponseWriter, r *http.Request) {
	run, ok := s.Store.GetRun(r.PathValue("id"))
	if !ok {
		fail(w, http.StatusNotFound, "run not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

type proofRequest struct {
	Subject   string `json:"subject"`
	Threshold int64  `json:"threshold"`
	Days      int    `json:"expiresInDays"`
}

func (s *Server) CreateProof(w http.ResponseWriter, r *http.Request) {
	var req proofRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Threshold <= 0 || req.Subject == "" {
		fail(w, http.StatusBadRequest, "invalid proof request")
		return
	}
	if req.Days <= 0 {
		req.Days = 30
	}
	p := s.Signer.NewProof(req.Subject, req.Threshold, req.Days)
	s.Store.PutProof(p)
	writeJSON(w, http.StatusCreated, map[string]any{"id": p.ID, "proof": p.Signature, "verifyUrl": "/proof/" + p.ID})
}

func (s *Server) GetProof(w http.ResponseWriter, r *http.Request) {
	p, ok := s.Store.GetProof(r.PathValue("id"))
	if !ok {
		fail(w, http.StatusNotFound, "proof not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// Portfolio reports real saved data. No balance is invented: edj does not
// hold funds, so the amounts are the totals of created requests and drafts.
func (s *Server) Portfolio(w http.ResponseWriter, r *http.Request) {
	links := s.Store.Links()
	runs := s.Store.Runs()
	proofs := s.Store.Proofs()
	var linkTotal, payrollTotal int64
	for _, l := range links {
		linkTotal += l.Amount
	}
	for _, run := range runs {
		payrollTotal += run.Total
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"links":       len(links),
		"payrollRuns": len(runs),
		"proofs":      len(proofs),
		"linkTotal":   linkTotal,
		"payrollTotal": payrollTotal,
	})
}
