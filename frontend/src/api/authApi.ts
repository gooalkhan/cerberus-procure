export interface User {
  id: number;
  username: string;
  display_name: string;
}

declare global {
  interface Window {
    login: (username: string, password: string) => Promise<string>;
  }
}

const isWasm = () => !!(window as any).login;

export const login = async (username: string, password: string): Promise<User> => {
  if (isWasm()) {
    try {
      const res = await window.login(username, password);
      return JSON.parse(res);
    } catch (e) {
      throw new Error(e as string);
    }
  }
  
  const res = await fetch('/api/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  });
  
  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || 'Login failed');
  }
  
  return res.json();
};
