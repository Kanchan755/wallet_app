package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kanchan755/wallet_app/monolith/internal/wallet/model"
)

type WalletRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, w *model.Wallet) error
	GetWalletByUserID(ctx context.Context, userID string) (*model.Wallet, error)
	UpdateBalanceTx(ctx context.Context, tx *sql.Tx, walletID string, newBalance float64, currentVersion int) error
}

type mysqlWalletRepository struct {
	db *sql.DB
}

func NewMySQLWalletRepository(db *sql.DB) WalletRepository {
	return &mysqlWalletRepository{db: db}
}

func (r *mysqlWalletRepository) CreateTx(ctx context.Context, tx *sql.Tx, w *model.Wallet) error {
	query := `
		INSERT INTO wallets (id, user_id, balance, currency, status, version)
		VALUES (?, ?, ?, ?, 1)
	`
	_, err := tx.ExecContext(ctx, query, w.ID, w.UserID, w.Balance, w.Currency, w.Status)
	if err != nil {
		return err
	}
	return nil
}

func (r *mysqlWalletRepository) GetWalletByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	query := `
		SELECT id, user_id, balance, currency, status, version, created_at, updated_at
		FROM wallets
		WHERE user_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, userID)
	var w model.Wallet
	if err := row.Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.Status, &w.Version, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *mysqlWalletRepository) UpdateBalanceTx(ctx context.Context, tx *sql.Tx, walletID string, newBalance float64, currentVersion int) error {
	query := `
	UPDATE wallets
	SET balance = ?, version = version + 1
	WHERE id = ? AND version = ?
	`
	result, err := tx.ExecContext(ctx, query, newBalance, walletID, currentVersion)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// if zero rows affected , it means database version has changed (concurrency conflict)
	if rowsAffected == 0 {
		return errors.New("concurrenct update detected: version mismatched")
	}
	return nil
}
