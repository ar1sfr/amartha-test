# Amartha Transaction Reconciliation Service

Reconciles internal system transactions against multiple banks’ statements to identify matched pairs, missing transactions on either side, and amount discrepancies within a specified timeframe.

## Features

- CSV ingestion for system transactions and multiple bank statements
- Normalizes signs and dates:
  - System `DEBIT` → negative, `CREDIT` → positive
  - Bank amounts already signed; debits are negative
  - Converts system `transactionTime` to date using a configurable timezone
- Deterministic matching:
  - Exact match first (same sign/date/amount)
  - Fallback to minimal absolute amount difference within a configurable date window
- Summary output:
  - `total_processed`, `total_matched`, `total_unmatched`
  - Unmatched system transactions
  - Unmatched bank transactions grouped by bank file
  - `total_discrepancy` = sum of absolute amount differences over matched pairs

## Requirements

- Go 1.22+
- No external dependencies; standard library only

## Project Structure

amartha-test/
├─ cmd/
│  └─ reconcile/
│     └─ main.go
├─ internal/
│  ├─ core/
│  │  ├─ reconcile.go
│  │  └─ reconcile_test.go
│  ├─ ingest/
│  │  ├─ bankcsv.go
│  │  ├─ common.go
│  │  ├─ parser_test.go
│  │  └─ systemcsv.go
│  └─ model/
│     └─ types.go
└─ go.mod

## CSV Formats

- System transactions CSV
  - Headers: `trxID,amount,type,transactionTime`
  - Example:
    ```
    trxID,amount,type,transactionTime
    S1,100.00,CREDIT,2025-01-02T10:00:00Z
    S2,50.25,DEBIT,2025-01-02 23:59:59
    ```

- Bank statements CSV (per bank file)
  - Headers: `unique_identifier,amount,date`
  - Example:
    ```
    unique_identifier,amount,date
    A1,100.00,2025-01-02
    A2,-50.25,2025-01-02
    ```

## Build and Test

- Build:
  - go build ./...
- Test (no cache):
  - go test -v -count=1 ./...
  Use `-count=1` to force re-execution when tests read external CSV fixtures.
- Clear test cache if needed:
  - go clean -testcache

## Running the CLI

- Flags:
- `--system` path to system CSV
- `--bank` path to bank CSV (repeatable for multiple banks)
- `--start` reconciliation start date `YYYY-MM-DD`
- `--end` reconciliation end date `YYYY-MM-DD`
- `--tz` timezone name (default `UTC`, example `Asia/Jakarta`)
- `--date-window` allowed posting-day shift (default `0`, example `1`)

- Example:
 - go run ./cmd/reconcile
    - --system ./data/system.csv
    - --bank ./data/bank_a.csv
    - --bank ./data/bank_b.csv
    - --start 2025-01-01
    - --end 2025-01-31
    - --tz UTC
    - --date-window 1

 - Sample output:
```json
{
  "total_processed": 2,
  "total_matched": 2,
  "total_unmatched": 1,
  "unmatched_system": [],
  "unmatched_bank": {
    "bank_b.csv": [
      { "unique_identifier": "B1", "amount": "-49.00", "date": "2025-01-02" }
    ]
  },
  "total_discrepancy": "0.00"
}
