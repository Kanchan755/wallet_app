package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	customError "github.com/kanchan755/wallet_app/monolith/internal/errors"
	"github.com/kanchan755/wallet_app/monolith/internal/logger"
	"github.com/kanchan755/wallet_app/monolith/internal/wallet/model"
	"github.com/kanchan755/wallet_app/monolith/internal/wallet/repository"
	"github.com/redis/go-redis/v9"
)

type WalletService interface {
	GetWalletByUserID(ctx context.Context, UserID string) (*model.Wallet, error)
}

type walletService struct {
	walletRepo repository.WalletRepository
	rdb        *redis.Client
}

func NewWalletService(walletRepo repository.WalletRepository, rdb *redis.Client) WalletService {
	return &walletService{
		walletRepo: walletRepo,
		rdb:        rdb,
	}
}

func (s *walletService) GetWalletByUserID(ctx context.Context, UserID string) (*model.Wallet, error) {
	cacheKey := fmt.Sprintf("user:wallet:%s", UserID)

	//check if data exist in redis
	cachedVal, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if err != redis.Nil {
			//redis is down or has an issue , dont fail the request
			logger.Warn(ctx, "Redis error during cache lookup, falling back to MySQL", "error", err.Error(), UserID)
		}
		//cache miss or redis down
		logger.Info(ctx, "cache miss for waller, fetching from MySQL..", "user_id", UserID)
	} else {
		//cache hit , desireliaze JSON string to model.wallet stuct
		wallet := &model.Wallet{}
		if err := json.Unmarshal([]byte(cachedVal), wallet); err == nil {
			logger.Info(ctx, "cache hit for wallet, returning from redis", "user_id", UserID)
			return wallet, nil
		}

		logger.Warn(ctx, "Failed to deserialize cached wallet data, falling back to MySQL", "user_id", UserID, "error", err.Error())
	}

	wallet, err := s.walletRepo.GetWalletByUserID(ctx, UserID)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "WALLET NOT FOUND", "Wallet for the given user is not found.")
	}

	// save to redis for 5 min
	wBytes, err := json.Marshal(wallet)
	if err == nil {
		s.rdb.Set(ctx, cacheKey, wBytes, 5*time.Minute)
		logger.Info(ctx, "Wallet data cached successfully", "user_id", UserID, "ttl", "5 min")
	}

	return wallet, nil
}
