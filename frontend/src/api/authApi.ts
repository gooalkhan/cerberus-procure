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

export const getSession = async (): Promise<User | null> => {
  if (isWasm()) {
    try {
      const res = await (window as any).getSession();
      return JSON.parse(res);
    } catch (e) {
      return null;
    }
  }
  
  const res = await fetch('/api/me');
  if (!res.ok) return null;
  return res.json();
};

export const logout = async () => {
  if (isWasm()) {
    localStorage.removeItem('session_user');
  } else {
    document.cookie = "session_id=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
  }
  window.location.reload();
};
