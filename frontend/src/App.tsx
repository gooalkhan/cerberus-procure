import { useState, useEffect } from 'react'
import { getTodos, addTodo, toggleTodo, deleteTodo, Todo } from './api/todoApi'

function App() {
  const [todos, setTodos] = useState<Todo[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadTodos()
  }, [])

  const loadTodos = async () => {
    try {
      const data = await getTodos()
      setTodos(data || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  const handleAdd = async () => {
    if (!input) return
    await addTodo(input)
    setInput('')
    loadTodos()
  }

  const handleToggle = async (id: string) => {
    await toggleTodo(id)
    loadTodos()
  }

  const handleDelete = async (id: string) => {
    await deleteTodo(id)
    loadTodos()
  }

  if (loading) return <div className="loading">Loading...</div>

  return (
    <div className="container">
      <h1>Cerberus Go Todo</h1>
      <div className="input-group">
        <input 
          value={input} 
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
          placeholder="What needs to be done?"
        />
        <button onClick={handleAdd}>Add</button>
      </div>
      <ul className="todo-list">
        {todos.map(todo => (
          <li key={todo.id} className={todo.completed ? 'completed' : ''}>
            <span onClick={() => handleToggle(todo.id)}>
              {todo.completed ? '✅' : '⬜'} {todo.title}
            </span>
            <button className="delete-btn" onClick={() => handleDelete(todo.id)}>Delete</button>
          </li>
        ))}
      </ul>
    </div>
  )
}

export default App
