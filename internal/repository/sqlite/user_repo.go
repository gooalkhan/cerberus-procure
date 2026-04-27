package sqlite

import (
	 "cerberus-procure/internal/models"
	"database/sql"
)

type SQLiteUserRepository struct {
	db *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) (*SQLiteUserRepository, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		display_name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return nil, err
	}
	return &SQLiteUserRepository{db: db}, nil
}

func (r *SQLiteUserRepository) GetUserByUsername(username string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow("SELECT id, username, password_hash, display_name, created_at, updated_at FROM users WHERE username = ?", username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *SQLiteUserRepository) CreateUser(user *models.User) error {
	res, err := r.db.Exec("INSERT INTO users (username, password_hash, display_name) VALUES (?, ?, ?)",
		user.Username, user.PasswordHash, user.DisplayName)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	user.ID = int(id)
	return nil
}
