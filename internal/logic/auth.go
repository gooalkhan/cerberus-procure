package logic

import (
	"cerberus-go/internal/models"
	"cerberus-go/internal/repository"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	userRepo repository.UserRepository
}

func NewAuthUseCase(userRepo repository.UserRepository) *AuthUseCase {
	return &AuthUseCase{userRepo: userRepo}
}

// Login validates user credentials and returns the user if successful
func (uc *AuthUseCase) Login(username, password string) (*models.User, error) {
	user, err := uc.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	// Compare hashed password with provided password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

// Register creates a new user with a hashed password
func (uc *AuthUseCase) Register(username, password, displayName string) (*models.User, error) {
	// Check if user already exists
	existing, _ := uc.userRepo.GetUserByUsername(username)
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		DisplayName:  displayName,
	}

	err = uc.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
