package main

import (
	"cerberus-go/internal/logic"
	"cerberus-go/internal/repository/sqlite"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var frontendAssets embed.FS

var todoUC *logic.TodoUseCase
var authUC *logic.AuthUseCase

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := authUC.Login(input.Username, input.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getTodosHandler(w http.ResponseWriter, r *http.Request) {
	todos, err := todoUC.GetTodos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func addTodoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	todo, err := todoUC.AddTodo(input.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func toggleTodoHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := todoUC.ToggleTodo(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func deleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := todoUC.DeleteTodo(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	todoRepo, err := sqlite.NewSQLiteTodoRepository("todos.db")
	if err != nil {
		panic(err)
	}
	userRepo, err := sqlite.NewSQLiteUserRepository(todoRepo.DB())
	if err != nil {
		panic(err)
	}

	todoUC = logic.NewTodoUseCase(todoRepo)
	authUC = logic.NewAuthUseCase(userRepo)

	// Seed admin user if not exists
	authUC.Register("admin", "1234", "Administrator")

	mux := http.NewServeMux()

	// API 핸들러
	mux.HandleFunc("/api/login", corsMiddleware(loginHandler))
	mux.HandleFunc("/api/todos", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getTodosHandler(w, r)
		case http.MethodPost:
			addTodoHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/todos/toggle", corsMiddleware(toggleTodoHandler))
	mux.HandleFunc("/api/todos/delete", corsMiddleware(deleteTodoHandler))

	// 프론트엔드 정적 파일 서빙
	distFS, _ := fs.Sub(frontendAssets, "dist")
	mux.Handle("/", http.FileServer(http.FS(distFS)))

	fmt.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", mux)
}
