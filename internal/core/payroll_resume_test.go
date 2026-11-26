package core

import (
	"strings"
	"testing"
)

func sampleRun() *PayrollRun {
	return NewPayrollRun([]PayrollRow{
		{Name: "Ada", Address: "bitcoincash:qaaa", Amount: 1000},
		{Name: "Grace", Address: "bitcoincash:qbbb", Amount: 2000},
	})
}

func TestPayrollRunStartsPending(t *testing.T) {
	run := sampleRun()
	if run.Total != 3000 {
		t.Fatalf("total = %d", run.Total)
	}
	if len(run.Results) != 2 || run.Results[0].Status != "pending" {
		t.Fatalf("results not initialized: %+v", run.Results)
	}
	if run.Done() {
		t.Fatal("new run reported done")
	}
}

func TestRecordResultIsIdempotentForBroadcast(t *testing.T) {
	run := sampleRun()
	if !run.RecordResult(0, "broadcast", "txid-one", "") {
		t.Fatal("first result rejected")
	}
	if run.RecordResult(0, "broadcast", "txid-two", "") {
		t.Fatal("completed result was overwritten")
	}
	if run.Results[0].Txid != "txid-one" {
		t.Fatalf("txid changed to %s", run.Results[0].Txid)
	}
	if run.Done() {
		t.Fatal("run done with one pending payout")
	}
	if !run.RecordResult(1, "broadcast", "txid-two", "") {
		t.Fatal("second result rejected")
	}
	if !run.Done() {
		t.Fatal("run not done after all payouts")
	}
}

func TestRecordResultRejectsInvalidInput(t *testing.T) {
	run := sampleRun()
	if run.RecordResult(-1, "broadcast", "x", "") || run.RecordResult(99, "broadcast", "x", "") {
		t.Fatal("out-of-range index accepted")
	}
	if run.RecordResult(0, "unknown", "x", "") {
		t.Fatal("unknown status accepted")
	}
}

func TestExportPayrollCSV(t *testing.T) {
	run := sampleRun()
	_ = run.RecordResult(0, "broadcast", "txid-one", "")
	csv := ExportPayrollCSV(run)
	if !strings.HasPrefix(csv, "name,address,amount,status,txid,error\n") {
		t.Fatalf("missing header: %s", csv)
	}
	if !strings.Contains(csv, `"Ada","bitcoincash:qaaa","1000","broadcast","txid-one",""`) {
		t.Fatalf("missing completed row: %s", csv)
	}
	if !strings.Contains(csv, `"Grace","bitcoincash:qbbb","2000","pending"`) {
		t.Fatalf("missing pending row: %s", csv)
	}
}
