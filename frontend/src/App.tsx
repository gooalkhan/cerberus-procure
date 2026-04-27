import { useState, useEffect } from 'react'
import { getTodos, addTodo, toggleTodo, deleteTodo, Todo } from './api/todoApi'
import { User } from './api/authApi'
import Login from './components/Login'

function App() {
  const [user, setUser] = useState<User | null>(null)
  const [todos, setTodos] = useState<Todo[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Check if user is already logged in (could use localStorage or a checkAuth API)
    setLoading(false)
  }, [])

  useEffect(() => {
    if (user) {
      loadTodos()
    }
  }, [user])

  const loadTodos = async () => {
    try {
      const data = await getTodos()
      setTodos(data || [])
    } catch (e) {
      console.error(e)
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

  const handleLogout = () => {
    setUser(null)
    setTodos([])
  }

  if (loading) return <div className="loading">Loading...</div>

  if (!user) {
    return <Login onLogin={setUser} />
  }

  return (
    <div className="container">
      <header>
        <h1>Cerberus Go Todo</h1>
        <div>
          <span>Welcome, {user.display_name}</span>
          <button className="logout-btn" onClick={handleLogout}>Logout</button>
        </div>
      </header>
      
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
