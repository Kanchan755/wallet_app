package service

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kanchan755/wallet_app/monolith/internal/auth"
	customError "github.com/kanchan755/wallet_app/monolith/internal/errors"
	"github.com/kanchan755/wallet_app/monolith/internal/user/model"
	"github.com/kanchan755/wallet_app/monolith/internal/user/repository"
	walletModel "github.com/kanchan755/wallet_app/monolith/internal/wallet/model"
	walletRepo "github.com/kanchan755/wallet_app/monolith/internal/wallet/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	// Define methods for user-related operations
	Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	GetProfile(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
}

type userService struct {
	db         *sql.DB
	UserRepo   repository.UserRepository
	walletRepo walletRepo.WalletRepository
}

func NewUserService(db *sql.DB, UserRepo repository.UserRepository, walletRepo walletRepo.WalletRepository) UserService {
	return &userService{
		db:         db,
		UserRepo:   UserRepo,
		walletRepo: walletRepo,
	}
}

// Implement the methods defined in the UserService interface
func (s *userService) Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	// Implement user registration logic here
	// For example, you can create a new user and save it to the database using the repository
	//1. check if the email already exists
	existingUser, _ := s.UserRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, customError.NewAppError(http.StatusConflict, "Email already exists", "this email is already registered")
	}
	// hash the password with bcrypt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, customError.ErrInternalServerError
	}

	//2. create a new user
	newUser := &model.User{
		ID:           uuid.New().String(), // Implement a function to generate unique IDs
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hashedBytes),
	}

	//begin the transaction database
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, customError.ErrInternalServerError
	}

	// we should rollback if anything error or in the middle
	// save to 2 table, users and wallet , if one of them lets say wallet is failed , then
	//users table to rollback
	defer tx.Rollback()

	// store user to db with a tx connection
	if err := s.UserRepo.CreateTx(ctx, tx, newUser); err != nil {
		return nil, customError.ErrInternalServerError
	}
	//Create a wallet for the new user
	newWallet := &walletModel.Wallet{
		ID:       uuid.New().String(),
		UserID:   newUser.ID,
		Balance:  0,
		Currency: "INR",
		Status:   "ACTIVE",
	}
	if err := s.walletRepo.CreateTx(ctx, tx, newWallet); err != nil {
		return nil, customError.ErrInternalServerError
	}

	// commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, customError.ErrInternalServerError
	}
	return s.UserRepo.FindByID(ctx, newUser.ID)
}

func (s *userService) GetProfile(ctx context.Context, id string) (*model.User, error) {
	// Implement logic to retrieve user profile by ID
	return s.UserRepo.FindByID(ctx, id)
}

func (s *userService) UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error) {
	// Implement logic to update user profile
	user, err := s.UserRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.FullName = req.FullName
	if err := s.UserRepo.Update(ctx, user); err != nil {
		return nil, err

	}
	return s.UserRepo.FindByID(ctx, id)
}

func (s *userService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	// find by email
	user, err := s.UserRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, customError.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "wrong email or password ")
	}

	// verify the hash password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, customError.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "wrong email or password ")
	}

	// generate access token 15 minutes
	accessToken, err := auth.GenerateToken(user.ID, user.Email, 15*time.Minute)
	if err != nil {
		return nil, customError.ErrInternalServerError
	}

	// generate refresh token 7 days
	refreshToken, err := auth.GenerateToken(user.ID, user.Email, 7*24*time.Hour)
	if err != nil {
		return nil, customError.ErrInternalServerError
	}
	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}
