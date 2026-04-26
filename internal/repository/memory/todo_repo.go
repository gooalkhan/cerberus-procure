package memory

import (
	"cerberus-go/internal/models"
	"errors"
	"fmt"
	"time"
)

type MemoryTodoRepository struct {
	todos []models.Todo
}

func NewMemoryTodoRepository() *MemoryTodoRepository {
	return &MemoryTodoRepository{
		todos: []models.Todo{},
	}
}

func (r *MemoryTodoRepository) GetTodos() ([]models.Todo, error) {
	return r.todos, nil
}

func (r *MemoryTodoRepository) AddTodo(title string) (models.Todo, error) {
	newTodo := models.Todo{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Title:     title,
		Completed: false,
	}
	r.todos = append(r.todos, newTodo)
	return newTodo, nil
}

func (r *MemoryTodoRepository) ToggleTodo(id string) error {
	for i, t := range r.todos {
		if t.ID == id {
			r.todos[i].Completed = !r.todos[i].Completed
			return nil
		}
	}
	return errors.New("todo not found")
}

func (r *MemoryTodoRepository) DeleteTodo(id string) error {
	for i, t := range r.todos {
		if t.ID == id {
			r.todos = append(r.todos[:i], r.todos[i+1:]...)
			return nil
		}
	}
	return errors.New("todo not found")
}
