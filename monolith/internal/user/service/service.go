package service

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	customError "github.com/kanchan755/wallet_app/monolith/internal/errors"
	"github.com/kanchan755/wallet_app/monolith/internal/user/model"
	"github.com/kanchan755/wallet_app/monolith/internal/user/repository"
)

type UserService interface {
	// Define methods for user-related operations
	Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	GetProfile(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// Implement the methods defined in the UserService interface
func (s *userService) Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	// Implement user registration logic here
	// For example, you can create a new user and save it to the database using the repository
	//1. check if the email already exists
	existingUser, _ := s.repo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, customError.NewAppError(http.StatusConflict, "Email already exists", "this email is already registered")
	}

	//2. create a new user
	newUser := &model.User{
		ID:           uuid.New().String(), // Implement a function to generate unique IDs
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: req.Password,
	}

	//3. save the new user to the database
	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, customError.ErrInternalServerError
	}
	return s.repo.FindByID(ctx, newUser.ID)
}

func (s *userService) GetProfile(ctx context.Context, id string) (*model.User, error) {
	// Implement logic to retrieve user profile by ID
	return s.repo.FindByID(ctx, id)
}

func (s *userService) UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error) {
	// Implement logic to update user profile
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.FullName = req.FullName
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err

	}
	return s.repo.FindByID(ctx, id)
}
