package sqlite

import (
	"cerberus-go/internal/models"
	"database/sql"
	"github.com/google/uuid"

	_ "modernc.org/sqlite"
)

type SQLiteTodoRepository struct {
	db *sql.DB
}

func NewSQLiteTodoRepository(dbPath string) (*SQLiteTodoRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id TEXT PRIMARY KEY,
		title TEXT,
		completed BOOLEAN
	)`)
	if err != nil {
		return nil, err
	}

	return &SQLiteTodoRepository{db: db}, nil
}

func (r *SQLiteTodoRepository) DB() *sql.DB {
	return r.db
}

func (r *SQLiteTodoRepository) GetTodos() ([]models.Todo, error) {
	rows, err := r.db.Query("SELECT id, title, completed FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

func (r *SQLiteTodoRepository) AddTodo(title string) (models.Todo, error) {
	id := uuid.New().String()
	todo := models.Todo{
		ID:        id,
		Title:     title,
		Completed: false,
	}

	_, err := r.db.Exec("INSERT INTO todos (id, title, completed) VALUES (?, ?, ?)", todo.ID, todo.Title, todo.Completed)
	if err != nil {
		return models.Todo{}, err
	}

	return todo, nil
}

func (r *SQLiteTodoRepository) ToggleTodo(id string) error {
	_, err := r.db.Exec("UPDATE todos SET completed = NOT completed WHERE id = ?", id)
	return err
}

func (r *SQLiteTodoRepository) DeleteTodo(id string) error {
	_, err := r.db.Exec("DELETE FROM todos WHERE id = ?", id)
	return err
}
