package service

import (
	"context"
	"net/http"

	customError "github.com/kanchan755/wallet_app/monolith/internal/errors"
	"github.com/kanchan755/wallet_app/monolith/internal/wallet/model"
	"github.com/kanchan755/wallet_app/monolith/internal/wallet/repository"
)

type WalletService interface {
	GetWalletByUserID(ctx context.Context, UserID string) (*model.Wallet, error)
}

type walletService struct {
	walletRepo repository.WalletRepository
}

func NewWalletService(walletRepo repository.WalletRepository) WalletService {
	return &walletService{
		walletRepo: walletRepo,
	}
}

func (s *walletService) GetWalletByUserID(ctx context.Context, UserID string) (*model.Wallet, error) {
	w, err := s.walletRepo.GetWalletByUserID(ctx, UserID)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "WALLET NOT FOUND", "Wallet for the given user is not found.")
	}
	return w, nil
}
