package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"edj.cash/edj/internal/chain"
	"edj.cash/edj/internal/core"
	"edj.cash/edj/internal/store"
)

type Server struct {
	Chain    chain.Chain
	Store    store.Store
	Signer   *core.Signer
	Electrum *chain.Electrum
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
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "engine": "pagor"})
}

type linkRequest struct {
	Amount int64   `json:"amount"`
	Token  *string `json:"token"`
	Note   string  `json:"note"`
}

// freshAddress derives a new xpub child per link when an xpub is configured.
// Without an xpub it falls back to the operator's configured address, which is
// the v1 behaviour. It never reuses an address already assigned to a link.
func (s *Server) freshAddress(ctx context.Context, token *string) (string, uint32, error) {
	xpub := os.Getenv("EDJ_XPUB")
	tokenAware := false
	if token != nil {
		xpub = os.Getenv("EDJ_TOKEN_XPUB")
		tokenAware = true
	}
	if xpub == "" {
		addr, err := s.Chain.AddressFor(ctx, token)
		return addr, 0, err
	}
	used := make(map[string]bool)
	for _, existing := range s.Store.Links() {
		used[existing.Address] = true
	}
	for attempt := 0; attempt < 12; attempt++ {
		index, err := chain.RandomDerivationIndex()
		if err != nil {
			return "", 0, err
		}
		addr, err := chain.DeriveAddress(xpub, index, tokenAware)
		if err != nil {
			return "", 0, err
		}
		if !used[addr] {
			return addr, index, nil
		}
	}
	return "", 0, fmt.Errorf("could not derive an unused address")
}

func (s *Server) CreateLink(w http.ResponseWriter, r *http.Request) {
	var req linkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		fail(w, http.StatusBadRequest, "invalid amount")
		return
	}
	addr, index, err := s.freshAddress(r.Context(), req.Token)
	if err != nil {
		fail(w, http.StatusInternalServerError, "address generation failed")
		return
	}
	l := core.NewLinkWithIndex(addr, index, req.Amount, req.Token, req.Note)
	s.Store.PutLink(l)
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": l.ID, "claimUrl": "/l/" + l.ID, "address": l.Address,
		"derivationIndex": l.DerivationIndex, "status": l.Status,
	})
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

// PayrollResult records a wallet-reported payout outcome. It is deliberately
// not a signing endpoint: the wallet signs and broadcasts, the server only
// keeps a resume-safe record of what happened.
func (s *Server) PayrollResult(w http.ResponseWriter, r *http.Request) {
	run, ok := s.Store.GetRun(r.PathValue("id"))
	if !ok {
		fail(w, http.StatusNotFound, "run not found")
		return
	}
	var req struct {
		Index  int    `json:"index"`
		Status string `json:"status"`
		Txid   string `json:"txid"`
		Error  string `json:"error"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "invalid result")
		return
	}
	if !run.RecordResult(req.Index, req.Status, req.Txid, req.Error) {
		fail(w, http.StatusBadRequest, "invalid or already recorded result")
		return
	}
	if run.Done() {
		run.Status = "complete"
	} else if run.Status == "draft" {
		run.Status = "in progress"
	}
	s.Store.PutRun(run)
	writeJSON(w, http.StatusOK, map[string]any{
		"id": run.ID, "status": run.Status, "done": run.Done(), "result": run.Results[req.Index],
	})
}

func (s *Server) PayrollExport(w http.ResponseWriter, r *http.Request) {
	run, ok := s.Store.GetRun(r.PathValue("id"))
	if !ok {
		fail(w, http.StatusNotFound, "run not found")
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"pagor-payroll-"+run.ID+".csv\"")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, core.ExportPayrollCSV(run))
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

// Portfolio reports real saved data. No balance is invented: pagor does not
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
		"links":        len(links),
		"payrollRuns":  len(runs),
		"proofs":       len(proofs),
		"linkTotal":    linkTotal,
		"payrollTotal": payrollTotal,
	})
}

func (s *Server) WalletUtxos(w http.ResponseWriter, r *http.Request) {
	if s.Electrum == nil {
		fail(w, http.StatusServiceUnavailable, "electrum not configured")
		return
	}
	address := r.URL.Query().Get("address")
	if address == "" {
		fail(w, http.StatusBadRequest, "address is required")
		return
	}
	utxos, err := s.Electrum.GetUtxos(address)
	if err != nil {
		fail(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, utxos)
}

func (s *Server) WalletBroadcast(w http.ResponseWriter, r *http.Request) {
	if s.Electrum == nil {
		fail(w, http.StatusServiceUnavailable, "electrum not configured")
		return
	}
	var req struct {
		RawTx string `json:"rawTx"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RawTx == "" {
		fail(w, http.StatusBadRequest, "rawTx is required")
		return
	}
	txid, err := s.Electrum.Broadcast(req.RawTx)
	if err != nil {
		fail(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"txid": txid})
}
