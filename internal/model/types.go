package model

import "time"

type TxType int

const (
	Debit  TxType = -1
	Credit TxType = 1
)

func (t TxType) String() string {
	if t == Credit {
		return "CREDIT"
	}
	return "DEBIT"
}

type SystemTx struct {
	TrxID       string
	AmountCents int64
	Type        TxType
	Time        time.Time
	Date        time.Time
}

type BankTx struct {
	ID          string
	AmountCents int64
	Date        time.Time
	BankName    string
}

type Match struct {
	System    SystemTx
	Bank      BankTx
	DiffCents int64
}

type Summary struct {
	TotalProcessed        int
	TotalMatched          int
	TotalUnmatched        int
	TotalDiscrepancyCents int64
}
