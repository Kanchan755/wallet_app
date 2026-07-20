package repository

import (
	"context"
	"database/sql"

	"github.com/kanchan755/wallet_app/monolith/internal/transaction/model"
)

type TransactionRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, t *model.Transaction) error
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*model.Transaction, error)
	GetHistory(ctx context.Context, userID string, params *model.PaginationParams) ([]model.Transaction, int64, error)
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

func (r *mysqlTransactionRepository) GetHistory(ctx context.Context, userID string, params *model.PaginationParams) ([]model.Transaction, int64, error) {
	// counting total data for pagination meta
	countQuery := `
SELECT COUNT(*)
FROM transaction WHERE (sender_wallet_id) = ? OR reciever_wallet_id = ?
`
	var total int64
	var err error

	if params.Status != "" {
		countQuery = countQuery + "AND status = ?"
		err = r.db.QueryRowContext(ctx, countQuery, userID, userID, params.Status).Scan(&total)
	} else {
		err = r.db.QueryRowContext(ctx, countQuery, userID, userID).Scan(&total)
	}
	if err != nil {
		return nil, 0, err
	}

	//get the paginated data, use sort and order
	//important use whitelist for  sort and order to prevent sql injection

	sortColumn := "created_at"
	if params.Sort == "ammount" {
		sortColumn = "ammount"
	}

	sortOrder := "desc"
	if params.Order == "asc" {
		sortOrder = "ASC"
	}

	query := `
	SELECT id, sender_wallet_id, receiver_wallet_id, amount, description, idempotency_key, status, created_at
	FROM transactions
	WHERE sender_wallet_id = ? OR receiver_wallet_id = ?
	`
	offset := (params.Page - 1) * params.Limit
	if offset < 0 {
		offset = 0
	}

	var rows *sql.Rows
	if params.Status != "" {
		query += " AND status = ? ORDER BY " + sortColumn + " " + sortOrder + " LIMIT ? OFFSET ?"
		rows, err = r.db.QueryContext(ctx, query, userID, userID, params.Status, params.Limit, offset)
	} else {
		query += " ORDER BY " + sortColumn + " " + sortOrder + " LIMIT ? OFFSET ?"
		rows, err = r.db.QueryContext(ctx, query, userID, userID, params.Limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var t model.Transaction
		var sender sql.NullString
		var receiver sql.NullString
		if err := rows.Scan(&t.ID, &sender, &receiver, &t.Amount, &t.Description, &t.IdempotencyKey, &t.Status, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		if sender.Valid {
			t.SenderWalletID = sender.String
		}
		if receiver.Valid {
			t.ReceiverWalletID = receiver.String
		}
		transactions = append(transactions, t)
	}
	return transactions, total, nil
}
