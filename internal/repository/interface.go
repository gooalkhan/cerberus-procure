package repository

import  "cerberus-procure/internal/models"

type TodoRepository interface {
	GetTodos() ([]models.Todo, error)
	AddTodo(title string) (models.Todo, error)
	ToggleTodo(id string) error
	DeleteTodo(id string) error
}

type UserRepository interface {
	GetUserByUsername(username string) (*models.User, error)
	CreateUser(user *models.User) error
}
