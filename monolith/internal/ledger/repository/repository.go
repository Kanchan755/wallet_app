package repository

import (
	"context"
	"database/sql"

	"github.com/kanchan755/wallet_app/monolith/internal/ledger/model"
)

type LedgerRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, entry *model.Ledger) error
	GetBalanceWalletID(ctx context.Context, walletID string) (float64, error)
	GetEntryByWalletID(ctx context.Context, walletID string) ([]model.Ledger, error)
	Close() error
}

type LedgerRepositoryImpl struct {
	db *sql.DB
}

func NewLedgerRepository(db *sql.DB) LedgerRepository {
	return &LedgerRepositoryImpl{db: db}
}

func (l *LedgerRepositoryImpl) CreateTx(ctx context.Context, tx *sql.Tx, entry *model.Ledger) error {
	query := `
	INSERT INTO ledgers (id, wallet_id, transaction_id, entry_type, amount)
	VALUES(?, ?, ?, ?, ?	)`

	_, err := tx.ExecContext(ctx, query, entry.ID, entry.WalletID, entry.TransactionID, entry.EntryType, entry.Amount)
	return err
}

func (l *LedgerRepositoryImpl) GetBalanceWalletID(ctx context.Context, walletID string) (float64, error) {
	query := `
	SELECT 
	COALESCE(SUM(CASE WHEN entry_type = 'CREDIT' THEN amount ELSE 0 END), 0) - 
	COALESCE(SUM(CASE WHEN entry_type = 'DEBIT' THEN amount ELSE 0 END), 0) as balance
	FROM ledgers
	WHERE wallet_id = ?`

	var balance float64
	err := l.db.QueryRowContext(ctx, query, walletID).Scan(&balance)
	return balance, err
}

func (l *LedgerRepositoryImpl) GetEntryByWalletID(ctx context.Context, walletID string) ([]model.Ledger, error) {
	query := `
	SELECT id, wallet_id, transaction_id, entry_type, amount
	FROM ledgers
	WHERE wallet_id = ?
	ORDER BY created_at ASC`

	rows, err := l.db.QueryContext(ctx, query, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.Ledger
	for rows.Next() {
		var entry model.Ledger
		if err := rows.Scan(&entry.ID, &entry.WalletID, &entry.TransactionID, &entry.EntryType, &entry.Amount, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (l *LedgerRepositoryImpl) Close() error {
	return l.db.Close()
}
