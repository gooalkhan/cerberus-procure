package main

import (
	"cerberus-go/internal/logic"
	"cerberus-go/internal/repository/memory"
	"encoding/json"
	"syscall/js"
)

var todoUC *logic.TodoUseCase
var authUC *logic.AuthUseCase

func login(this js.Value, args []js.Value) interface{} {
	username := args[0].String()
	password := args[1].String()
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			user, err := authUC.Login(username, password)
			if err != nil {
				reject.Invoke(err.Error())
				return
			}
			b, _ := json.Marshal(user)
			resolve.Invoke(string(b))
		}()
		return nil
	})

	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func getTodos(this js.Value, args []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		// reject := promiseArgs[1]

		go func() {
			todos, _ := todoUC.GetTodos()
			b, _ := json.Marshal(todos)
			resolve.Invoke(string(b))
		}()
		return nil
	})

	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func addTodo(this js.Value, args []js.Value) interface{} {
	title := args[0].String()
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		go func() {
			todo, _ := todoUC.AddTodo(title)
			b, _ := json.Marshal(todo)
			resolve.Invoke(string(b))
		}()
		return nil
	})
	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func toggleTodo(this js.Value, args []js.Value) interface{} {
	id := args[0].String()
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		go func() {
			todoUC.ToggleTodo(id)
			resolve.Invoke()
		}()
		return nil
	})
	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func deleteTodo(this js.Value, args []js.Value) interface{} {
	id := args[0].String()
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		go func() {
			todoUC.DeleteTodo(id)
			resolve.Invoke()
		}()
		return nil
	})
	promiseClass := js.Global().Get("Promise")
	return promiseClass.New(handler)
}

func main() {
	todoRepo := memory.NewMemoryTodoRepository()
	userRepo := memory.NewMemoryUserRepository()
	
	todoUC = logic.NewTodoUseCase(todoRepo)
	authUC = logic.NewAuthUseCase(userRepo)

	// Seed admin user
	authUC.Register("admin", "1234", "Administrator")

	js.Global().Set("login", js.FuncOf(login))
	js.Global().Set("getTodos", js.FuncOf(getTodos))
	js.Global().Set("addTodo", js.FuncOf(addTodo))
	js.Global().Set("toggleTodo", js.FuncOf(toggleTodo))
	js.Global().Set("deleteTodo", js.FuncOf(deleteTodo))

	select {}
}
