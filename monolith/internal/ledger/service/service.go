package service

import (
	"context"

	"github.com/kanchan755/wallet_app/monolith/internal/ledger/model"
	"github.com/kanchan755/wallet_app/monolith/internal/ledger/repository"
	walletRepo "github.com/kanchan755/wallet_app/monolith/internal/wallet/repository"
)

type LedgerService interface {
	ReconcileWalletBalance(ctx context.Context, userID string) (bool, float64, float64, error)
	GetMutationHistory(ctx context.Context, userID string) ([]model.Ledger, error)
}

type ledgerServiceImpl struct {
	lRepo repository.LedgerRepository
	wRepo walletRepo.WalletRepository
}

func NewLedgerService(lRepo repository.LedgerRepository, wRepo walletRepo.WalletRepository) LedgerService {
	return &ledgerServiceImpl{
		lRepo: lRepo,
		wRepo: wRepo,
	}
}

func (s *ledgerServiceImpl) ReconcileWalletBalance(ctx context.Context, userID string) (bool, float64, float64, error) {
	//get the user wallet data
	wallet, err := s.wRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return false, 0, 0, err
	}

	//get the sum of all enteries
	ledgerBalance, err := s.lRepo.GetBalanceWalletID(ctx, wallet.ID)
	if err != nil {
		return false, 0, 0, err
	}

	//check if there is a discrepancy
	isConsistent := wallet.Balance == ledgerBalance
	if !isConsistent {
		//update the wallet balance

	}

	return isConsistent, wallet.Balance, ledgerBalance, nil
}

func (s *ledgerServiceImpl) GetMutationHistory(ctx context.Context, userID string) ([]model.Ledger, error) {
	wallet, err := s.wRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.lRepo.GetEntryByWalletID(ctx, wallet.ID)
}
