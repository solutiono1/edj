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

// Supabase stores records through the PostgREST API using snake_case column
// names that match supabase/schema.sql. The service role key stays on the
// backend and is never sent to the browser.
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

type linkRow struct {
	ID              string    `json:"id"`
	Address         string    `json:"address"`
	Amount          int64     `json:"amount"`
	Token           *string   `json:"token"`
	Note            string    `json:"note"`
	CreatedAt       time.Time `json:"created_at"`
	Status          string    `json:"status"`
	DerivationIndex uint32    `json:"derivation_index"`
	OwnerSub        string    `json:"owner_sub"`
}

func (s *Supabase) PutLink(l *core.Link) {
	body := map[string]any{
		"id":               l.ID,
		"address":          l.Address,
		"amount":           l.Amount,
		"token":            l.Token,
		"note":             l.Note,
		"created_at":       l.CreatedAt,
		"status":           l.Status,
		"derivation_index": l.DerivationIndex,
		"owner_sub":        l.OwnerSub,
	}
	if err := s.request(http.MethodPost, "links", url.Values{"on_conflict": {"id"}}, body, nil); err != nil {
		fmt.Println("PutLink:", err)
	}
}

func (s *Supabase) GetLink(id string) (*core.Link, bool) {
	var out []linkRow
	if err := s.request(http.MethodGet, "links", url.Values{"id": {"eq." + id}}, nil, &out); err != nil || len(out) == 0 {
		return nil, false
	}
	return &core.Link{
		ID: out[0].ID, Address: out[0].Address, Amount: out[0].Amount,
		Token: out[0].Token, Note: out[0].Note, CreatedAt: out[0].CreatedAt, Status: out[0].Status,
		DerivationIndex: out[0].DerivationIndex,
		OwnerSub:        out[0].OwnerSub,
	}, true
}

func (s *Supabase) Links() []*core.Link {
	var out []linkRow
	if err := s.request(http.MethodGet, "links", url.Values{"select": {"*"}}, nil, &out); err != nil {
		return nil
	}
	links := make([]*core.Link, 0, len(out))
	for _, row := range out {
		links = append(links, &core.Link{
			ID: row.ID, Address: row.Address, Amount: row.Amount,
			Token: row.Token, Note: row.Note, CreatedAt: row.CreatedAt, Status: row.Status,
			DerivationIndex: row.DerivationIndex,
			OwnerSub:        row.OwnerSub,
		})
	}
	return links
}

type payrollRow struct {
	ID          string              `json:"id"`
	RowsJSON    []core.PayrollRow   `json:"rows_json"`
	Total       int64               `json:"total"`
	PlanJSON    []core.PayoutStep   `json:"plan_json"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	ResultsJSON []core.PayoutResult `json:"results_json"`
	OwnerSub    string              `json:"owner_sub"`
}

func (r payrollRow) toRun() *core.PayrollRun {
	return &core.PayrollRun{ID: r.ID, Rows: r.RowsJSON, Total: r.Total, Plan: r.PlanJSON, Status: r.Status, Results: r.ResultsJSON, OwnerSub: r.OwnerSub}
}

func (s *Supabase) PutRun(r *core.PayrollRun) {
	body := map[string]any{
		"id": r.ID, "rows_json": r.Rows, "total": r.Total,
		"plan_json": r.Plan, "status": r.Status, "created_at": time.Now().UTC(),
		"results_json": r.Results,
		"owner_sub":    r.OwnerSub,
	}
	_ = s.request(http.MethodPost, "payroll_runs", url.Values{"on_conflict": {"id"}}, body, nil)
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

type proofRow struct {
	ID                string            `json:"id"`
	Subject           string            `json:"subject"`
	Threshold         int64             `json:"threshold"`
	ExpiresAt         time.Time         `json:"expires_at"`
	Signature         string            `json:"signature"`
	PublicKey         string            `json:"public_key"`
	Claim             string            `json:"claim"`
	OwnerSub          string            `json:"owner_sub"`
	SignerFingerprint string            `json:"signer_fingerprint"`
	Anchor            *core.ProofAnchor `json:"anchor"`
}

type privatePayrollRow struct {
	ID         string    `json:"id"`
	OwnerSub   string    `json:"owner_sub"`
	Salt       string    `json:"salt"`
	IV         string    `json:"iv"`
	Ciphertext string    `json:"ciphertext"`
	RowCount   int       `json:"row_count"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

func (s *Supabase) PutPrivatePayroll(p *core.PrivatePayroll) {
	_ = s.request(http.MethodPost, "private_payrolls", url.Values{"on_conflict": {"id"}}, map[string]any{
		"id": p.ID, "owner_sub": p.OwnerSub, "salt": p.Salt, "iv": p.IV,
		"ciphertext": p.Ciphertext, "row_count": p.RowCount, "status": p.Status,
		"created_at": p.CreatedAt,
	}, nil)
}

func (s *Supabase) GetPrivatePayroll(id string) (*core.PrivatePayroll, bool) {
	var out []privatePayrollRow
	if err := s.request(http.MethodGet, "private_payrolls", url.Values{"id": {"eq." + id}}, nil, &out); err != nil || len(out) == 0 {
		return nil, false
	}
	return &core.PrivatePayroll{
		ID: out[0].ID, OwnerSub: out[0].OwnerSub, Salt: out[0].Salt, IV: out[0].IV,
		Ciphertext: out[0].Ciphertext, RowCount: out[0].RowCount, Status: out[0].Status,
		CreatedAt: out[0].CreatedAt,
	}, true
}

func (s *Supabase) PutProof(p *core.DisclosureProof) {
	body := map[string]any{
		"id": p.ID, "subject": p.Subject, "threshold": p.Threshold,
		"expires_at": p.ExpiresAt, "signature": p.Signature,
		"public_key": p.PublicKey, "claim": p.Claim,
		"owner_sub": p.OwnerSub, "signer_fingerprint": p.SignerFingerprint,
		"anchor": p.Anchor,
	}
	_ = s.request(http.MethodPost, "proofs", url.Values{"on_conflict": {"id"}}, body, nil)
}

func (s *Supabase) GetProof(id string) (*core.DisclosureProof, bool) {
	var out []proofRow
	if err := s.request(http.MethodGet, "proofs", url.Values{"id": {"eq." + id}}, nil, &out); err != nil || len(out) == 0 {
		return nil, false
	}
	return &core.DisclosureProof{
		ID: out[0].ID, Subject: out[0].Subject, Threshold: out[0].Threshold,
		ExpiresAt: out[0].ExpiresAt, Signature: out[0].Signature,
		PublicKey: out[0].PublicKey, Claim: out[0].Claim, OwnerSub: out[0].OwnerSub,
		SignerFingerprint: out[0].SignerFingerprint, Anchor: out[0].Anchor,
	}, true
}

func (s *Supabase) Proofs() []*core.DisclosureProof {
	var out []proofRow
	if err := s.request(http.MethodGet, "proofs", url.Values{"select": {"*"}}, nil, &out); err != nil {
		return nil
	}
	proofs := make([]*core.DisclosureProof, 0, len(out))
	for _, row := range out {
		proofs = append(proofs, &core.DisclosureProof{
			ID: row.ID, Subject: row.Subject, Threshold: row.Threshold,
			ExpiresAt: row.ExpiresAt, Signature: row.Signature,
			PublicKey: row.PublicKey, Claim: row.Claim, OwnerSub: row.OwnerSub,
			SignerFingerprint: row.SignerFingerprint, Anchor: row.Anchor,
		})
	}
	return proofs
}

// userRow includes the password hash, which is deliberately not part of User
// because User is serialized to the browser by /api/me.
type userRow struct {
	Sub          string     `json:"sub"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	PasswordHash string     `json:"password_hash"`
	Status       string     `json:"status"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

func (r userRow) user() *User {
	return &User{Sub: r.Sub, Email: r.Email, Name: r.Name, Status: r.Status, DeletedAt: r.DeletedAt}
}

func (u *User) rowWithHash(hash string) userRow {
	return userRow{
		Sub: u.Sub, Email: u.Email, Name: u.Name, PasswordHash: hash,
		Status: u.Status, DeletedAt: u.DeletedAt,
	}
}

func (s *Supabase) UpsertUser(u *User) (*User, bool) {
	var existing []userRow
	if err := s.request(http.MethodGet, "users", url.Values{"sub": {"eq." + u.Sub}}, nil, &existing); err != nil {
		return u, false
	}
	hash := ""
	status := ""
	var deletedAt *time.Time
	if len(existing) > 0 {
		hash = existing[0].PasswordHash
		status = existing[0].Status
		deletedAt = existing[0].DeletedAt
	}
	if status == "" {
		status = "active"
	}
	u.Status = status
	u.DeletedAt = deletedAt
	row := userRow{
		Sub: u.Sub, Email: u.Email, Name: u.Name, PasswordHash: hash,
		Status: status, DeletedAt: deletedAt,
	}
	_ = s.request(http.MethodPost, "users", url.Values{"on_conflict": {"sub"}}, row, nil)
	return u, len(existing) == 0
}

func (s *Supabase) User(sub string) (*User, bool) {
	var out []userRow
	if err := s.request(http.MethodGet, "users", url.Values{"sub": {"eq." + sub}}, nil, &out); err != nil || len(out) == 0 {
		return nil, false
	}
	return out[0].user(), true
}

func (s *Supabase) UserByEmail(email string) (*User, bool) {
	var out []userRow
	if err := s.request(http.MethodGet, "users", url.Values{"email": {"eq." + email}}, nil, &out); err != nil || len(out) == 0 {
		return nil, false
	}
	return out[0].user(), true
}

func (s *Supabase) CreateEmailUser(u *User, hash string) (*User, bool) {
	var bySub []userRow
	if err := s.request(http.MethodGet, "users", url.Values{"sub": {"eq." + u.Sub}}, nil, &bySub); err != nil || len(bySub) > 0 {
		return nil, false
	}
	var byEmail []userRow
	if err := s.request(http.MethodGet, "users", url.Values{"email": {"eq." + u.Email}}, nil, &byEmail); err != nil || len(byEmail) > 0 {
		return nil, false
	}
	if u.Status == "" {
		u.Status = "active"
	}
	err := s.request(http.MethodPost, "users", url.Values{"on_conflict": {"sub"}}, u.rowWithHash(hash), nil)
	if err != nil {
		return nil, false
	}
	return u, true
}

func (s *Supabase) PasswordHash(sub string) (string, bool) {
	var out []userRow
	q := url.Values{"sub": {"eq." + sub}, "select": {"password_hash"}}
	if err := s.request(http.MethodGet, "users", q, nil, &out); err != nil || len(out) == 0 || out[0].PasswordHash == "" {
		return "", false
	}
	return out[0].PasswordHash, true
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

func (s *Supabase) UpdateUser(u *User) {
	var existing []userRow
	hash := ""
	if err := s.request(http.MethodGet, "users", url.Values{"sub": {"eq." + u.Sub}}, nil, &existing); err == nil && len(existing) > 0 {
		hash = existing[0].PasswordHash
	}
	_ = s.request(http.MethodPost, "users", url.Values{"on_conflict": {"sub"}}, u.rowWithHash(hash), nil)
}

func (s *Supabase) DeleteSessionsForSub(sub string) {
	var sessions []struct {
		Token string `json:"token"`
	}
	if err := s.request(http.MethodGet, "sessions", url.Values{"sub": {"eq." + sub}}, nil, &sessions); err != nil {
		return
	}
	for _, session := range sessions {
		s.DeleteSession(session.Token)
	}
}

func (s *Supabase) PurgeAccountsBefore(before time.Time) int {
	var expired []userRow
	q := url.Values{
		"status":     {"eq.deleted"},
		"deleted_at": {"lt." + before.UTC().Format(time.RFC3339Nano)},
	}
	if err := s.request(http.MethodGet, "users", q, nil, &expired); err != nil {
		return 0
	}
	for _, row := range expired {
		s.DeleteSessionsForSub(row.Sub)
		for _, table := range []string{"links", "payroll_runs", "proofs", "private_payrolls"} {
			_ = s.request(http.MethodDelete, table, url.Values{"owner_sub": {"eq." + row.Sub}}, nil, nil)
		}
		_ = s.request(http.MethodDelete, "users", url.Values{"sub": {"eq." + row.Sub}}, nil, nil)
	}
	return len(expired)
}

type notificationRow struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
}

func (s *Supabase) Notifications() []Notification {
	var out []notificationRow
	if err := s.request(http.MethodGet, "notifications", url.Values{"order": {"created_at.asc"}}, nil, &out); err != nil {
		return nil
	}
	notifs := make([]Notification, 0, len(out))
	for _, row := range out {
		notifs = append(notifs, Notification{
			ID: row.ID, Title: row.Title, Body: row.Body, CreatedAt: row.CreatedAt, Read: row.Read,
		})
	}
	return notifs
}
