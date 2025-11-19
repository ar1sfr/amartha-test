package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"amartha-test/internal/core"
	"amartha-test/internal/ingest"
	"amartha-test/internal/model"
)

type stringSlice []string

func (s *stringSlice) String() string     { return strings.Join(*s, ",") }
func (s *stringSlice) Set(v string) error { *s = append(*s, v); return nil }

type SystemTxOut struct {
	TrxID  string `json:"trxID"`
	Amount string `json:"amount"`
	Type   string `json:"type"`
	Date   string `json:"date"`
}

type BankTxOut struct {
	UniqueIdentifier string `json:"unique_identifier"`
	Amount           string `json:"amount"`
	Date             string `json:"date"`
}

func formatCents(c int64) string {
	if c < 0 {
		c = -c
		return "-" + fmt.Sprintf("%d.%02d", c/100, c%100)
	}
	return fmt.Sprintf("%d.%02d", c/100, c%100)
}

func mapSystemOut(xs []model.SystemTx) []SystemTxOut {
	r := make([]SystemTxOut, len(xs))
	for i, x := range xs {
		r[i] = SystemTxOut{
			TrxID:  x.TrxID,
			Amount: formatCents(x.AmountCents),
			Type:   x.Type.String(),
			Date:   x.Date.Format("2006-01-02"),
		}
	}
	return r
}

func mapBankOut(m map[string][]model.BankTx) map[string][]BankTxOut {
	out := make(map[string][]BankTxOut)

	for k, xs := range m {
		r := make([]BankTxOut, len(xs))
		for i, x := range xs {
			r[i] = BankTxOut{
				UniqueIdentifier: x.ID,
				Amount:           formatCents(x.AmountCents),
				Date:             x.Date.Format("2006-01-02"),
			}
		}
		out[k] = r
	}
	return out
}

func main() {
	var systemPath string
	var bankPaths stringSlice
	var startStr string
	var endStr string
	var tzStr string
	var dateWindow int

	flag.StringVar(&systemPath, "system", "", "system transactions CSV path")
	flag.Var(&bankPaths, "bank", "bank statement CSV path (repeatable)")
	flag.StringVar(&startStr, "start", "", "start date YYYY-MM-DD")
	flag.StringVar(&endStr, "end", "", "end date YYYY-MM-DD")
	flag.StringVar(&tzStr, "tz", "UTC", "timezone name")
	flag.IntVar(&dateWindow, "date-window", 0, "date window in days")
	flag.Parse()

	if systemPath == "" || len(bankPaths) == 0 || startStr == "" || endStr == "" {
		fmt.Fprintln(os.Stderr, "missing required flags: --system, --bank, --start, --end")
		os.Exit(1)
	}

	loc, err := time.LoadLocation(tzStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	startDate, err := time.ParseInLocation("2006-01-02", startStr, loc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	endDate, err := time.ParseInLocation("2006-01-02", endStr, loc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sys, err := ingest.ParseSystemCSV(systemPath, loc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var banks []model.BankTx
	for _, p := range bankPaths {
		name := filepath.Base(p)
		b, err := ingest.ParseBankCSV(p, name, loc)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		banks = append(banks, b...)
	}

	summary, _, unmatchedSystem, unmatchedBank := core.Reconcile(sys, banks, startDate, endDate, dateWindow)
	out := struct {
		TotalProcessed   int                    `json:"total_processed"`
		TotalMatched     int                    `json:"total_matched"`
		TotalUnmatched   int                    `json:"total_unmatched"`
		UnmatchedSystem  []SystemTxOut          `json:"unmatched_system"`
		UnmatchedBank    map[string][]BankTxOut `json:"unmatched_bank"`
		TotalDiscrepancy string                 `json:"total_discrepancy"`
	}{
		TotalProcessed:   summary.TotalProcessed,
		TotalMatched:     summary.TotalMatched,
		TotalUnmatched:   summary.TotalUnmatched,
		UnmatchedSystem:  mapSystemOut(unmatchedSystem),
		UnmatchedBank:    mapBankOut(unmatchedBank),
		TotalDiscrepancy: formatCents(summary.TotalDiscrepancyCents),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
