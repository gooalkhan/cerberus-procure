package logic

import (
	 "cerberus-procure/internal/models"
	 "cerberus-procure/internal/repository"
)

type TodoUseCase struct {
	repo repository.TodoRepository
}

func NewTodoUseCase(repo repository.TodoRepository) *TodoUseCase {
	return &TodoUseCase{repo: repo}
}

func (uc *TodoUseCase) GetTodos() ([]models.Todo, error) {
	return uc.repo.GetTodos()
}

func (uc *TodoUseCase) AddTodo(title string) (models.Todo, error) {
	return uc.repo.AddTodo(title)
}

func (uc *TodoUseCase) ToggleTodo(id string) error {
	return uc.repo.ToggleTodo(id)
}

func (uc *TodoUseCase) DeleteTodo(id string) error {
	return uc.repo.DeleteTodo(id)
}
