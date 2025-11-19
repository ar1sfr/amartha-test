package ingest

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"amartha-test/internal/model"
)

func ParseBankCSV(path string, bankName string, loc *time.Location) ([]model.BankTx, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	hdr, err := r.Read()
	if err != nil {
		return nil, err
	}

	idx := map[string]int{}
	for i, h := range hdr {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	req := []string{"unique_identifier", "amount", "date"}
	for _, k := range req {
		if _, ok := idx[k]; !ok {
			return nil, errors.New("missing header " + k + " in" + filepath.Base(path))
		}
	}

	var out []model.BankTx
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		id := strings.TrimSpace(rec[idx["unique_identifier"]])
		amtStr := strings.TrimSpace(rec[idx["amount"]])
		dateStr := strings.TrimSpace(rec[idx["date"]])
		if id == "" {
			continue
		}
		cents, err := parseCents(amtStr)
		if err != nil {
			return nil, err
		}
		d, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			return nil, err
		}

		out = append(out, model.BankTx{
			ID:          id,
			AmountCents: cents,
			Date:        d,
			BankName:    bankName,
		})
	}
	return out, nil
}
