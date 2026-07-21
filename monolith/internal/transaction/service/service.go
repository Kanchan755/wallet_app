package service

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	customError "github.com/kanchan755/wallet_app/monolith/internal/errors"
	ledgerModel "github.com/kanchan755/wallet_app/monolith/internal/ledger/model"
	ledgerRepo "github.com/kanchan755/wallet_app/monolith/internal/ledger/repository"
	"github.com/kanchan755/wallet_app/monolith/internal/logger"
	txModel "github.com/kanchan755/wallet_app/monolith/internal/transaction/model"
	"github.com/kanchan755/wallet_app/monolith/internal/transaction/repository"
	userRepo "github.com/kanchan755/wallet_app/monolith/internal/user/repository"
	walletRepo "github.com/kanchan755/wallet_app/monolith/internal/wallet/repository"
	"github.com/redis/go-redis/v9"
)

type TransactionService interface {
	Transfer(ctx context.Context, senderUserID string, req txModel.TransferRequest) (*txModel.Transaction, error)
	GetHistory(ctx context.Context, userID string, params *txModel.PaginationParams) ([]txModel.Transaction, *txModel.PaginationMeta, error)
}

type transactionServiceImpl struct {
	db         *sql.DB
	txRepo     repository.TransactionRepository
	userRepo   userRepo.UserRepository
	walletRepo walletRepo.WalletRepository
	ledgerRepo ledgerRepo.LedgerRepository
	rdb        *redis.Client
}

func NewTransactionService(
	db *sql.DB,
	rdb *redis.Client,
	txRepo repository.TransactionRepository,
	userRepo userRepo.UserRepository,
	walletRepo walletRepo.WalletRepository,
	ledgerRepo ledgerRepo.LedgerRepository,
) TransactionService {
	return &transactionServiceImpl{
		db:         db,
		rdb:        rdb,
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

	transaction.CreatedAt = time.Now()

	//invalidate cache
	senderCachekey := "wallet:user:" + senderUserID
	receiverCachekey := "wallet:user:" + recieverUser.ID

	//delete the cache keys asynchronously (dont block http response)
	go func() {
		logger.Info(ctx, "Deleting cache asynchronously", "sender_id", senderUserID, "receiver_id", recieverUser.ID)
		s.rdb.Del(context.Background(), senderCachekey)
		s.rdb.Del(context.Background(), receiverCachekey)
	}()
	return transaction, nil
}

func (s *transactionServiceImpl) GetHistory(ctx context.Context, userID string, params *txModel.PaginationParams) ([]txModel.Transaction, *txModel.PaginationMeta, error) {

	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, nil, customError.NewAppError(http.StatusNotFound, "WALLET_NOT_FOUND", "Wallet not found")
	}

	// max limit
	if params.Limit > 100 {
		params.Limit = 100
	}

	transaction, total, err := s.txRepo.GetHistory(ctx, wallet.ID, params)

	if err != nil {
		return nil, nil, customError.NewAppError(http.StatusInternalServerError, "DB_ERROR", "Failed to get transaction history")
	}

	totalPage := int(total / int64(params.Limit))
	if total%int64(params.Limit) != 0 {
		totalPage++
	}
	meta := &txModel.PaginationMeta{
		Page:      params.Page,
		Limit:     params.Limit,
		TotalData: total,
		TotalPage: totalPage,
	}

	return transaction, meta, nil
}
