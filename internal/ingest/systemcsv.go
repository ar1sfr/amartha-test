package ingest

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"amartha-test/internal/model"
)

func ParseSystemCSV(path string, loc *time.Location) ([]model.SystemTx, error) {
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

	req := []string{"trxid", "amount", "type", "transactiontime"}
	for _, k := range req {
		if _, ok := idx[k]; !ok {
			return nil, errors.New("missing header " + k)
		}
	}

	var out []model.SystemTx
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		trxID := strings.TrimSpace(rec[idx["trxid"]])
		amtStr := strings.TrimSpace(rec[idx["amount"]])
		typStr := strings.ToUpper(strings.TrimSpace(rec[idx["type"]]))
		timeStr := strings.TrimSpace(rec[idx["transactiontime"]])
		if trxID == "" {
			continue
		}
		cents, err := parseCents(amtStr)
		if err != nil {
			return nil, err
		}

		var t model.TxType
		switch typStr {
		case "CREDIT":
			t = model.Credit
		case "DEBIT":
			t = model.Debit
		default:
			return nil, errors.New("invalid type " + typStr)
		}

		tm, err := parseAnyTime(timeStr, loc)
		if err != nil {
			return nil, err
		}
		d := time.Date(tm.In(loc).Year(), tm.In(loc).Month(), tm.In(loc).Day(), 0, 0, 0, 0, loc)
		if t == model.Debit && cents > 0 {
			cents = -cents
		}
		if t == model.Credit && cents < 0 {
			cents = -cents
		}

		out = append(out, model.SystemTx{
			TrxID:       trxID,
			AmountCents: cents,
			Type:        t,
			Time:        tm.In(loc),
			Date:        d,
		})
	}
	return out, nil
}
