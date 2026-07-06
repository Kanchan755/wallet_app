package model

import "time"

type Ledger struct {
	ID            string    `json:"id"`
	WalletID      string    `json:"wallet_id"`
	TransactionID string    `json:"transaction_id"`
	EntryType     string    `json:"entry_type"`
	Amount        float64   `json:"amount"`
	BalanceAfter  float64   `json:"balance_after"`
	CreatedAt     time.Time `json:"created_at"`
}
