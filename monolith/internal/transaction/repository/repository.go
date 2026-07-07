package repository

import (
	"context"
	"database/sql"

	"github.com/kanchan755/wallet_app/monolith/internal/transaction/model"
)

type TransactionRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, t *model.Transaction) error
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*model.Transaction, error)
}

type mysqlTransactionRepository struct {
	db *sql.DB
}

func NewMySQLTransactionRepository(db *sql.DB) TransactionRepository {
	return &mysqlTransactionRepository{db: db}
}

func (r *mysqlTransactionRepository) CreateTx(ctx context.Context, tx *sql.Tx, t *model.Transaction) error {
	query := `
	INSERT INTO transactions (id, sender_wallet_id, receiver_wallet_id, amount, description, idempotency_key, status)
	VALUES(?, ?, ?, ?, ?, ?, ?)
	`
	_, err := tx.ExecContext(ctx, query, t.ID, t.SenderWalletID, t.ReceiverWalletID, t.Amount, t.Description, t.IdempotencyKey, t.Status)
	return err
}

func (r *mysqlTransactionRepository) GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*model.Transaction, error) {
	query := `
	SELECT id, sender_wallet_id, receiver_wallet_id, amount, description, idempotency_key, status, created_at
	FROM transactions
	WHERE idempotency_key = ?
	`
	var t model.Transaction
	var sender sql.NullString
	var receiver sql.NullString
	if err := r.db.QueryRowContext(ctx, query, idempotencyKey).Scan(&t.ID, &sender, &receiver, &t.Amount, &t.Description, &t.IdempotencyKey, &t.Status, &t.CreatedAt); err != nil {
		return nil, err
	}
	if sender.Valid {
		t.SenderWalletID = sender.String
	}
	if receiver.Valid {
		t.ReceiverWalletID = receiver.String
	}
	return &t, nil
}
