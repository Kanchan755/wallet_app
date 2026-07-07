package service

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	customError "github.com/kanchan755/wallet_app/monolith/internal/errors"
	ledgerModel "github.com/kanchan755/wallet_app/monolith/internal/ledger/model"
	ledgerRepo "github.com/kanchan755/wallet_app/monolith/internal/ledger/repository"
	txModel "github.com/kanchan755/wallet_app/monolith/internal/transaction/model"
	"github.com/kanchan755/wallet_app/monolith/internal/transaction/repository"
	userRepo "github.com/kanchan755/wallet_app/monolith/internal/user/repository"
	walletRepo "github.com/kanchan755/wallet_app/monolith/internal/wallet/repository"
)

type TransactionService interface {
	Transfer(ctx context.Context, senderUserID string, req txModel.TransferRequest) (*txModel.Transaction, error)
}

type transactionServiceImpl struct {
	db         *sql.DB
	txRepo     repository.TransactionRepository
	userRepo   userRepo.UserRepository
	walletRepo walletRepo.WalletRepository
	ledgerRepo ledgerRepo.LedgerRepository
}

func NewTransactionService(
	db *sql.DB,
	txRepo repository.TransactionRepository,
	userRepo userRepo.UserRepository,
	walletRepo walletRepo.WalletRepository,
	ledgerRepo ledgerRepo.LedgerRepository,
) TransactionService {
	return &transactionServiceImpl{
		db:         db,
		txRepo:     txRepo,
		userRepo:   userRepo,
		walletRepo: walletRepo,
		ledgerRepo: ledgerRepo,
	}
}

func (s *transactionServiceImpl) Transfer(ctx context.Context, senderUserID string, req txModel.TransferRequest) (*txModel.Transaction, error) {
	// check idempotency key (this is for checking to not to reprocess the same request )
	existingTx, err := s.txRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return existingTx, nil
	}

	//look reciever by email
	recieverUser, err := s.userRepo.FindByEmail(ctx, req.RecieverEmail)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "RECIEVER_NOT_FOUND", "Reciever not found")
	}
	// start db transaction

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, customError.ErrInternalServerError
	}
	// use go lang defer to rollback the transaction if something goes wrong
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	// look for sender and reciever wallet
	senderWallet, err := s.walletRepo.GetWalletByUserID(ctx, senderUserID)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "SENDER_NOT_FOUND", "Sender not found")
	}
	recieverWallet, err := s.walletRepo.GetWalletByUserID(ctx, recieverUser.ID)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "RECIEVER_NOT_FOUND", "Reciever not found")
	}
	// sender and reciever cannot be same
	if senderWallet.ID == recieverWallet.ID {
		return nil, customError.NewAppError(http.StatusBadRequest, "INVALID_TRANSFER", "Sender and reciever cannot be same")
	}
	// validate sender wallet is enough or not
	if senderWallet.Balance < req.Amount {
		return nil, customError.NewAppError(http.StatusBadRequest, "INSUFFICIENT_FUNDS", "Insufficient funds")
	}
	// reduce sender wallet and reciever wallet with tx and checking version
	newSenderBalance := senderWallet.Balance - req.Amount
	err = s.walletRepo.UpdateBalanceTx(ctx, tx, senderWallet.ID, newSenderBalance, senderWallet.Version+1)
	if err != nil {
		//if failed because of version mismaych , return special error so client can retry
		return nil, customError.NewAppError(http.StatusConflict, "CONCURRENCY_CONFLICT", "Transaction is busy , please retry in the followinf minutes ")

	}
	newRecieverBalance := recieverWallet.Balance + req.Amount
	err = s.walletRepo.UpdateBalanceTx(ctx, tx, recieverWallet.ID, newRecieverBalance, recieverWallet.Version+1)
	if err != nil {
		//if failed because of version mismaych , return special error so client can retry
		return nil, customError.NewAppError(http.StatusConflict, "CONCURRENCY_CONFLICT", "Transaction is busy , please retry in the followinf minutes ")
	}

	// create data transaction record
	// create transaction record
	transaction := &txModel.Transaction{
		ID:               uuid.New().String(),
		SenderWalletID:   senderWallet.ID,
		ReceiverWalletID: recieverWallet.ID,
		Amount:           req.Amount,
		Description:      req.Description,
		IdempotencyKey:   req.IdempotencyKey,
		Status:           "SUCCESS",
	}
	if err := s.txRepo.CreateTx(ctx, tx, transaction); err != nil {
		return nil, customError.NewAppError(http.StatusInternalServerError, "DB_ERROR", "Failed to create transaction")
	}
	// sender wallet ledger (debit)
	senderLedger := &ledgerModel.Ledger{
		ID:            uuid.New().String(),
		WalletID:      senderWallet.ID,
		TransactionID: transaction.ID,
		EntryType:     "DEBIT",
		Amount:        req.Amount,
		BalanceAfter:  newSenderBalance,
	}
	if err := s.ledgerRepo.CreateTx(ctx, tx, senderLedger); err != nil {
		return nil, customError.NewAppError(http.StatusInternalServerError, "DB_ERROR", "Failed to create transaction")
	}
	// reciever wallet ledger (credit)
	recieverLedger := &ledgerModel.Ledger{
		ID:            uuid.New().String(),
		WalletID:      recieverWallet.ID,
		TransactionID: transaction.ID,
		EntryType:     "CREDIT",
		Amount:        req.Amount,
		BalanceAfter:  newRecieverBalance,
	}
	if err := s.ledgerRepo.CreateTx(ctx, tx, recieverLedger); err != nil {
		return nil, customError.NewAppError(http.StatusInternalServerError, "DB_ERROR", "Failed to create transaction")
	}

	// commit the db transaction
	if err := tx.Commit(); err != nil {
		return nil, customError.NewAppError(http.StatusInternalServerError, "DB_ERROR", "Failed to commit transaction")
	}

	return transaction, nil
}
