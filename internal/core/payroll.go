package core

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type PayrollRow struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

type PayoutStep struct {
	Kind string `json:"kind"` // swap | mix | transfer
	Note string `json:"note"`
}

type PayoutResult struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
	Status  string `json:"status"` // pending | signed | broadcast | failed
	Txid    string `json:"txid,omitempty"`
	Error   string `json:"error,omitempty"`
}

type PayrollRun struct {
	ID      string         `json:"id"`
	Rows    []PayrollRow   `json:"rows"`
	Total   int64          `json:"total"`
	Plan    []PayoutStep   `json:"plan"`
	Status  string         `json:"status"`
	Results []PayoutResult `json:"results"`
}

// ParsePayroll strictly parses name,address,amount CSV. No formula evaluation.
func ParsePayroll(r io.Reader) ([]PayrollRow, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = 3
	cr.TrimLeadingSpace = true
	records, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid csv: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty csv")
	}
	header := records[0]
	if strings.ToLower(header[0]) != "name" || strings.ToLower(header[1]) != "address" || strings.ToLower(header[2]) != "amount" {
		return nil, fmt.Errorf("expected header name,address,amount")
	}
	rows := make([]PayrollRow, 0, len(records)-1)
	for _, rec := range records[1:] {
		amt, err := strconv.ParseInt(strings.TrimSpace(rec[2]), 10, 64)
		if err != nil || amt <= 0 {
			return nil, fmt.Errorf("invalid amount for %s", rec[0])
		}
		rows = append(rows, PayrollRow{Name: strings.TrimSpace(rec[0]), Address: strings.TrimSpace(rec[1]), Amount: amt})
	}
	return rows, nil
}

// NewPayrollRun plans an unlinkable payout route. Honest: steps are descriptive,
// not a claim of anonymity. v1 executes via wallet + CashFusion externally.
func NewPayrollRun(rows []PayrollRow) *PayrollRun {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	total := int64(0)
	results := make([]PayoutResult, 0, len(rows))
	for i, r := range rows {
		total += r.Amount
		results = append(results, PayoutResult{
			Index: i, Name: r.Name, Address: r.Address, Amount: r.Amount, Status: "pending",
		})
	}
	return &PayrollRun{
		ID:      "PR_" + hex.EncodeToString(b),
		Rows:    rows,
		Total:   total,
		Plan:    []PayoutStep{{Kind: "swap", Note: "PUSD to BCH for privacy route"}, {Kind: "mix", Note: "CashFusion on BCH leg"}, {Kind: "transfer", Note: "one-time output per recipient"}},
		Status:  "draft",
		Results: results,
	}
}

// RecordResult applies a wallet-reported payout outcome. A completed result is
// never overwritten, so a retry after a crash cannot double-record a payment.
func (r *PayrollRun) RecordResult(index int, status, txid, errText string) bool {
	if index < 0 || index >= len(r.Results) {
		return false
	}
	current := &r.Results[index]
	if current.Status == "broadcast" && current.Txid != "" {
		return false
	}
	switch status {
	case "signed", "broadcast", "failed":
	default:
		return false
	}
	current.Status = status
	if txid != "" {
		current.Txid = txid
	}
	if errText != "" {
		current.Error = errText
	}
	return true
}

// Done reports whether every payout has reached a terminal broadcast state.
func (r *PayrollRun) Done() bool {
	if len(r.Results) == 0 {
		return false
	}
	for _, result := range r.Results {
		if result.Status != "broadcast" {
			return false
		}
	}
	return true
}

// ExportPayrollCSV writes the current payout outcomes as CSV.
func ExportPayrollCSV(run *PayrollRun) string {
	var b strings.Builder
	b.WriteString("name,address,amount,status,txid,error\n")
	for _, result := range run.Results {
		fields := []string{
			result.Name, result.Address, strconv.FormatInt(result.Amount, 10),
			result.Status, result.Txid, result.Error,
		}
		for i, field := range fields {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(`"` + strings.ReplaceAll(field, `"`, `""`) + `"`)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
