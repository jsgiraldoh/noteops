import { api, setToken, clearToken } from './client';

export interface User { id: string; full_name: string; email: string; role: string; }
export interface LoginResponse { token: string; user: User; }

export async function login(email: string, password: string): Promise<LoginResponse> {
  const res = await api.post<LoginResponse>('/auth/login', { email, password });
  setToken(res.token);
  if (typeof localStorage !== 'undefined') localStorage.setItem('noteops_token', res.token);
  return res;
}

export function logout() {
  clearToken();
  if (typeof localStorage !== 'undefined') localStorage.removeItem('noteops_token');
}

export function restoreToken() {
  if (typeof localStorage !== 'undefined') {
    const t = localStorage.getItem('noteops_token');
    if (t) setToken(t);
    return t;
  }
  return null;
}
