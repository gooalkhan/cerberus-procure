package memory

import (
	 "cerberus-procure/internal/models"
	"errors"
	"sync"
	"time"
)

type MemoryUserRepository struct {
	users map[string]*models.User
	mu    sync.RWMutex
	nextID int
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:  make(map[string]*models.User),
		nextID: 1,
	}
}

func (r *MemoryUserRepository) GetUserByUsername(username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[username]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (r *MemoryUserRepository) CreateUser(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.Username]; ok {
		return errors.New("user already exists")
	}
	
	user.ID = r.nextID
	r.nextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	r.users[user.Username] = user
	return nil
}
