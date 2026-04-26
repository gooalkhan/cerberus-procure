export interface Todo {
  id: string;
  title: string;
  completed: boolean;
}

declare global {
  interface Window {
    Go: any;
    getTodos: () => Promise<string>;
    addTodo: (title: string) => Promise<string>;
    toggleTodo: (id: string) => Promise<void>;
    deleteTodo: (id: string) => Promise<void>;
  }
}

const isWasm = () => !!(window as any).getTodos;

export const getTodos = async (): Promise<Todo[]> => {
  if (isWasm()) {
    const res = await window.getTodos();
    return JSON.parse(res);
  }
  const res = await fetch('/api/todos');
  return res.json();
};

export const addTodo = async (title: string): Promise<Todo> => {
  if (isWasm()) {
    const res = await window.addTodo(title);
    return JSON.parse(res);
  }
  const res = await fetch('/api/todos', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title }),
  });
  return res.json();
};

export const toggleTodo = async (id: string): Promise<void> => {
  if (isWasm()) {
    await window.toggleTodo(id);
    return;
  }
  await fetch(`/api/todos/toggle?id=${id}`, { method: 'POST' });
};

export const deleteTodo = async (id: string): Promise<void> => {
  if (isWasm()) {
    await window.deleteTodo(id);
    return;
  }
  await fetch(`/api/todos/delete?id=${id}`, { method: 'POST' });
};
