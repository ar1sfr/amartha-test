package core

import (
	"testing"
	"time"

	"amartha-test/internal/model"
)

func TestExactMatch(t *testing.T) {
	loc := time.UTC
	sys := []model.SystemTx{
		{TrxID: "A", AmountCents: 10000, Type: model.Credit, Time: time.Date(2025, 1, 2, 10, 0, 0, 0, loc), Date: time.Date(2025, 1, 2, 0, 0, 0, 0, loc)},
	}
	banks := []model.BankTx{
		{ID: "X1", AmountCents: 10000, Date: time.Date(2025, 1, 2, 0, 0, 0, 0, loc), BankName: "bankA.csv"},
	}
	res, matches, unmatchedSys, unmatchedBank := Reconcile(sys, banks, time.Date(2025, 1, 1, 0, 0, 0, 0, loc), time.Date(2025, 1, 3, 0, 0, 0, 0, loc), 0)
	if res.TotalProcessed != 1 || res.TotalMatched != 1 || res.TotalUnmatched != 0 || res.TotalDiscrepancyCents != 0 {
		t.Fatalf("unexpected summary %+v", res)
	}
	if len(matches) != 1 || matches[0].DiffCents != 0 {
		t.Fatalf("unexpected matches %+v", matches)
	}
	if len(unmatchedSys) != 0 {
		t.Fatalf("unexpected unmatched system %d", len(unmatchedSys))
	}
	if len(unmatchedBank["bankA.csv"]) != 0 {
		t.Fatalf("unexpected unmatched bank %+v", unmatchedBank)
	}
}

func TestAmountDiscrepancy(t *testing.T) {
	loc := time.UTC
	sys := []model.SystemTx{
		{TrxID: "A", AmountCents: 10000, Type: model.Credit, Time: time.Date(2025, 1, 2, 10, 0, 0, 0, loc), Date: time.Date(2025, 1, 2, 0, 0, 0, 0, loc)},
	}
	banks := []model.BankTx{
		{ID: "X1", AmountCents: 9900, Date: time.Date(2025, 1, 2, 0, 0, 0, 0, loc), BankName: "bankA.csv"},
	}
	res, matches, _, _ := Reconcile(sys, banks, time.Date(2025, 1, 1, 0, 0, 0, 0, loc), time.Date(2025, 1, 3, 0, 0, 0, 0, loc), 0)
	if res.TotalMatched != 1 || res.TotalDiscrepancyCents != 100 {
		t.Fatalf("unexpected result %+v", res)
	}
	if matches[0].DiffCents != 100 {
		t.Fatalf("unexpected diff %d", matches[0].DiffCents)
	}
}

func TestSystemOnlyUnmatched(t *testing.T) {
	loc := time.UTC
	sys := []model.SystemTx{
		{TrxID: "A", AmountCents: -5000, Type: model.Debit, Time: time.Date(2025, 1, 2, 10, 0, 0, 0, loc), Date: time.Date(2025, 1, 2, 0, 0, 0, 0, loc)},
	}
	banks := []model.BankTx{}
	res, _, unmatchedSys, unmatchedBank := Reconcile(sys, banks, time.Date(2025, 1, 1, 0, 0, 0, 0, loc), time.Date(2025, 1, 3, 0, 0, 0, 0, loc), 0)
	if res.TotalMatched != 0 || res.TotalUnmatched != 1 {
		t.Fatalf("unexpected result %+v", res)
	}
	if len(unmatchedSys) != 1 || len(unmatchedBank) != 0 {
		t.Fatalf("unexpected unmatched %d %d", len(unmatchedSys), len(unmatchedBank))
	}
}

func TestBankOnlyUnmatchedGrouped(t *testing.T) {
	loc := time.UTC
	sys := []model.SystemTx{}
	banks := []model.BankTx{
		{ID: "B1", AmountCents: 10000, Date: time.Date(2025, 1, 2, 0, 0, 0, 0, loc), BankName: "bankA.csv"},
		{ID: "B2", AmountCents: -2000, Date: time.Date(2025, 1, 2, 0, 0, 0, 0, loc), BankName: "bankB.csv"},
	}
	res, _, unmatchedSys, unmatchedBank := Reconcile(sys, banks, time.Date(2025, 1, 1, 0, 0, 0, 0, loc), time.Date(2025, 1, 3, 0, 0, 0, 0, loc), 0)
	if res.TotalMatched != 0 || res.TotalUnmatched != 2 {
		t.Fatalf("unexpected result %+v", res)
	}
	if len(unmatchedSys) != 0 {
		t.Fatalf("unexpected unmatched system %d", len(unmatchedSys))
	}
	if len(unmatchedBank["bankA.csv"]) != 1 || len(unmatchedBank["bankB.csv"]) != 1 {
		t.Fatalf("unexpected unmatched bank %+v", unmatchedBank)
	}
}

func TestDateWindow(t *testing.T) {
	loc := time.UTC
	sys := []model.SystemTx{
		{TrxID: "A", AmountCents: 10000, Type: model.Credit, Time: time.Date(2025, 1, 2, 10, 0, 0, 0, loc), Date: time.Date(2025, 1, 2, 0, 0, 0, 0, loc)},
	}
	banks := []model.BankTx{
		{ID: "X1", AmountCents: 10000, Date: time.Date(2025, 1, 3, 0, 0, 0, 0, loc), BankName: "bankA.csv"},
	}
	res, matches, _, _ := Reconcile(sys, banks, time.Date(2025, 1, 2, 0, 0, 0, 0, loc), time.Date(2025, 1, 3, 0, 0, 0, 0, loc), 1)
	if res.TotalMatched != 1 || len(matches) != 1 {
		t.Fatalf("unexpected result %+v", res)
	}
}
