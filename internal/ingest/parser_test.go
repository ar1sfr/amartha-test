package ingest

import (
	"os"
	"testing"
	"time"
)

func TestParseSystemCSV(t *testing.T) {
	loc := time.UTC
	tmp := t.TempDir()
	path := tmp + "/system.csv"
	data := "trxID,amount,type,transactionTime\nS1,100.00,CREDIT,2025-01-02T10:00:00Z\nS2,50.00,DEBIT,2025-01-02 23:59:59\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	sys, err := ParseSystemCSV(path, loc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(sys) != 2 {
		t.Fatalf("len %d", len(sys))
	}
	if sys[0].AmountCents != 10000 || sys[0].Type.String() != "CREDIT" {
		t.Fatalf("wrong %+v", sys[0])
	}
	if sys[1].AmountCents != -5000 || sys[1].Type.String() != "DEBIT" {
		t.Fatalf("wrong %+v", sys[1])
	}
}

func TestParseBankCSV(t *testing.T) {
	loc := time.UTC
	tmp := t.TempDir()
	path := tmp + "/bank.csv"
	data := "unique_identifier,amount,date\nB1,100.00,2025-01-02\nB2,-50.00,2025-01-02\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	banks, err := ParseBankCSV(path, "bank.csv", loc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(banks) != 2 {
		t.Fatalf("len %d", len(banks))
	}
	if banks[0].AmountCents != 10000 || banks[0].BankName != "bank.csv" {
		t.Fatalf("wrong %+v", banks[0])
	}
	if banks[1].AmountCents != -5000 {
		t.Fatalf("wrong %+v", banks[1])
	}
}
