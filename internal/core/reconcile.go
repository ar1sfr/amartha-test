package core

import (
	"time"

	"amartha-test/internal/model"
)

type bankIndex struct {
	txs        []model.BankTx
	used       []bool
	byDateSign map[string]map[int][]int
}

func buildBankIndex(txs []model.BankTx) *bankIndex {
	idx := &bankIndex{
		txs:        txs,
		used:       make([]bool, len(txs)),
		byDateSign: make(map[string]map[int][]int),
	}

	for i, x := range txs {
		key := x.Date.Format("2006-01-02")
		sign := 1
		if x.AmountCents < 0 {
			sign = -1
		}
		m, ok := idx.byDateSign[key]
		if !ok {
			m = make(map[int][]int)
			idx.byDateSign[key] = m
		}
		m[sign] = append(m[sign], i)
	}
	return idx
}

func candidateIndexes(idx *bankIndex, date time.Time, sign int, window int) []int {
	var out []int

	for d := -window; d <= window; d++ {
		dd := date.AddDate(0, 0, d).Format("2006-01-02")
		ms := idx.byDateSign[dd]
		if ms == nil {
			continue
		}
		out = append(out, ms[sign]...)
	}
	return out
}

func Reconcile(systemTxs []model.SystemTx, bankTxs []model.BankTx, startDate, endDate time.Time, dateWindow int) (model.Summary, []model.Match, []model.SystemTx, map[string][]model.BankTx) {
	var sys []model.SystemTx
	for _, x := range systemTxs {
		if !x.Date.Before(startDate) && !x.Date.After(endDate) {
			sys = append(sys, x)
		}
	}

	var bankFiltered []model.BankTx
	for _, x := range bankTxs {
		if !x.Date.Before(startDate) && !x.Date.After(endDate) {
			bankFiltered = append(bankFiltered, x)
		}
	}

	groups := map[string][]model.BankTx{}
	for _, x := range bankFiltered {
		groups[x.BankName] = append(groups[x.BankName], x)
	}

	byBank := map[string]*bankIndex{}
	for name, xs := range groups {
		byBank[name] = buildBankIndex(xs)
	}

	var matches []model.Match
	var unmatchedSystem []model.SystemTx
	for _, s := range sys {
		sign := 1
		if s.AmountCents < 0 {
			sign = -1
		}

		var bestBank string
		var bestIdx int = -1
		var bestDiff int64 = 1 << 62
		var exactBank string
		var exactIdx int = -1
		for bankName, idx := range byBank {
			cand := candidateIndexes(idx, s.Date, sign, dateWindow)
			for _, i := range cand {
				if idx.used[i] {
					continue
				}
				diff := s.AmountCents - idx.txs[i].AmountCents
				if diff == 0 {
					exactBank = bankName
					exactIdx = i
					break
				}
				if diff < 0 {
					diff = -diff
				}
				if diff < bestDiff {
					bestDiff = diff
					bestIdx = i
					bestBank = bankName
				}
			}
			if exactIdx != -1 {
				break
			}
		}
		if exactIdx != -1 {
			b := byBank[exactBank]
			b.used[exactIdx] = true
			matches = append(matches, model.Match{System: s, Bank: b.txs[exactIdx], DiffCents: 0})
			continue
		}
		if bestIdx != -1 {
			b := byBank[bestBank]
			b.used[bestIdx] = true
			matches = append(matches, model.Match{System: s, Bank: b.txs[bestIdx], DiffCents: bestDiff})
		} else {
			unmatchedSystem = append(unmatchedSystem, s)
		}
	}

	unmatchedBank := map[string][]model.BankTx{}
	var totalDiscrepancy int64
	for _, m := range matches {
		if m.DiffCents < 0 {
			totalDiscrepancy += -m.DiffCents
		} else {
			totalDiscrepancy += m.DiffCents
		}
	}
	for bankName, idx := range byBank {
		for i := range idx.txs {
			if !idx.used[i] {
				unmatchedBank[bankName] = append(unmatchedBank[bankName], idx.txs[i])
			}
		}
	}

	totalUnmatched := len(unmatchedSystem)
	for _, xs := range unmatchedBank {
		totalUnmatched += len(xs)
	}
	summary := model.Summary{
		TotalProcessed:        len(sys),
		TotalMatched:          len(matches),
		TotalUnmatched:        totalUnmatched,
		TotalDiscrepancyCents: totalDiscrepancy,
	}
	return summary, matches, unmatchedSystem, unmatchedBank
}
