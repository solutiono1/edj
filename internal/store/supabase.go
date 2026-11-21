package store

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"edj.cash/edj/internal/core"
)

// Supabase stores records through the PostgREST API. It uses the service role
// key, which stays on the backend and is never sent to the browser.
type Supabase struct {
	baseURL string
	key     string
	client  *http.Client
}

func NewSupabase(baseURL, key string) (*Supabase, *core.Signer, error) {
	s := &Supabase{
		baseURL: baseURL,
		key:     key,
		client:  &http.Client{Timeout: 10 * time.Second},
	}

	var meta []struct {
		Value string `json:"value"`
	}
	if err := s.request(http.MethodGet, "meta", url.Values{"key": {"eq.signer_key"}}, nil, &meta); err != nil {
		return nil, nil, err
	}

	var signer *core.Signer
	if len(meta) > 0 && meta[0].Value != "" {
		raw, err := base64.RawURLEncoding.DecodeString(meta[0].Value)
		if err != nil {
			return nil, nil, fmt.Errorf("decode stored signing key: %w", err)
		}
		signer, err = core.NewSignerFromKey(ed25519.PrivateKey(raw))
		if err != nil {
			return nil, nil, fmt.Errorf("load stored signing key: %w", err)
		}
	} else {
		signer = core.NewSigner()
		encoded := base64.RawURLEncoding.EncodeToString(signer.PrivateKey())
		if err := s.request(http.MethodPost, "meta", url.Values{"on_conflict": {"key"}}, map[string]string{
			"key": "signer_key", "value": encoded,
		}, nil); err != nil {
			return nil, nil, err
		}
	}
	return s, signer, nil
}

func (s *Supabase) request(method, table string, query url.Values, body any, out any) error {
	endpoint := s.baseURL + "/rest/v1/" + table
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", s.key)
	req.Header.Set("Authorization", "Bearer "+s.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates,return=representation")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("supabase %s %s: %s", method, table, string(data))
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

func (s *Supabase) PutLink(l *core.Link) {
	_ = s.request(http.MethodPost, "links", url.Values{"on_conflict": {"id"}}, l, nil)
}
func (s *Supabase) GetLink(id string) (*core.Link, bool) {
	var out []*core.Link
	if err := s.request(http.MethodGet, "links", url.Values{"id": {"eq." + id}}, nil, &out); err != nil || len(out) == 0 {
		return nil, false
	}
	return out[0], true
}
func (s *Supabase) Links() []*core.Link {
	var out []*core.Link
	if err := s.request(http.MethodGet, "links", url.Values{"select": {"*"}}, nil, &out); err != nil {
		return nil
	}
	return out
}

func (s *Supabase) PutRun(r *core.PayrollRun) {
	body := map[string]any{
		"id": r.ID, "rows_json": r.Rows, "total": r.Total,
		"plan_json": r.Plan, "status": r.Status, "created_at": time.Now().UTC(),
	}
	_ = s.request(http.MethodPost, "payroll_runs", url.Values{"on_conflict": {"id"}}, body, nil)
}
type payrollRow struct {
	ID       string             `json:"id"`
	RowsJSON []core.PayrollRow  `json:"rows_json"`
	Total    int64              `json:"total"`
	PlanJSON []core.PayoutStep  `json:"plan_json"`
	Status   string             `json:"status"`
}

func (r payrollRow) toRun() *core.PayrollRun {
	return &core.PayrollRun{ID: r.ID, Rows: r.RowsJSON, Total: r.Total, Plan: r.PlanJSON, Status: r.Status}
}

func (s *Supabase) GetRun(id string) (*core.PayrollRun, bool) {
	var out []payrollRow
	if err := s.request(http.MethodGet, "payroll_runs", url.Values{"id": {"eq." + id}}, nil, &out); err != nil || len(out) == 0 {
		return nil, false
	}
	return out[0].toRun(), true
}
func (s *Supabase) Runs() []*core.PayrollRun {
	var out []payrollRow
	if err := s.request(http.MethodGet, "payroll_runs", url.Values{"select": {"*"}}, nil, &out); err != nil {
		return nil
	}
	runs := make([]*core.PayrollRun, 0, len(out))
	for _, row := range out {
		runs = append(runs, row.toRun())
	}
	return runs
}

func (s *Supabase) PutProof(p *core.DisclosureProof) {
	body := map[string]any{
		"id": p.ID, "subject": p.Subject, "threshold": p.Threshold,
		"expires_at": p.ExpiresAt, "signature": p.Signature,
		"public_key": p.PublicKey, "claim": p.Claim,
	}
	_ = s.request(http.MethodPost, "proofs", url.Values{"on_conflict": {"id"}}, body, nil)
}
func (s *Supabase) GetProof(id string) (*core.DisclosureProof, bool) {
	var out []*core.DisclosureProof
	if err := s.request(http.MethodGet, "proofs", url.Values{"id": {"eq." + id}}, nil, &out); err != nil || len(out) == 0 {
		return nil, false
	}
	return out[0], true
}
func (s *Supabase) Proofs() []*core.DisclosureProof {
	var out []*core.DisclosureProof
	if err := s.request(http.MethodGet, "proofs", url.Values{"select": {"*"}}, nil, &out); err != nil {
		return nil
	}
	return out
}

func (s *Supabase) UpsertUser(u *User) (*User, bool) {
	var existing []*User
	if err := s.request(http.MethodGet, "users", url.Values{"sub": {"eq." + u.Sub}}, nil, &existing); err != nil {
		return u, false
	}
	_ = s.request(http.MethodPost, "users", url.Values{"on_conflict": {"sub"}}, u, nil)
	return u, len(existing) == 0
}
func (s *Supabase) User(sub string) (*User, bool) {
	var out []*User
	if err := s.request(http.MethodGet, "users", url.Values{"sub": {"eq." + sub}}, nil, &out); err != nil || len(out) == 0 {
		return nil, false
	}
	return out[0], true
}

func (s *Supabase) PutSession(token, sub string) {
	_ = s.request(http.MethodPost, "sessions", url.Values{"on_conflict": {"token"}}, map[string]string{
		"token": token, "sub": sub,
	}, nil)
}
func (s *Supabase) Session(token string) (string, bool) {
	var out []struct {
		Sub string `json:"sub"`
	}
	if err := s.request(http.MethodGet, "sessions", url.Values{"token": {"eq." + token}}, nil, &out); err != nil || len(out) == 0 {
		return "", false
	}
	return out[0].Sub, true
}
func (s *Supabase) DeleteSession(token string) {
	_ = s.request(http.MethodDelete, "sessions", url.Values{"token": {"eq." + token}}, nil, nil)
}

func (s *Supabase) Notifications() []Notification {
	var out []Notification
	if err := s.request(http.MethodGet, "notifications", url.Values{"order": {"created_at.asc"}}, nil, &out); err != nil {
		return nil
	}
	return out
}
