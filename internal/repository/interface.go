package repository

import "cerberus-go/internal/models"

type TodoRepository interface {
	GetTodos() ([]models.Todo, error)
	AddTodo(title string) (models.Todo, error)
	ToggleTodo(id string) error
	DeleteTodo(id string) error
}
