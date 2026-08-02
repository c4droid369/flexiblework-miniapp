// Token storage helpers. Kept tiny on purpose — request.ts owns the actual
// Authorization header construction so the API surface is centralized.

const ACCESS_KEY = 'admin_template_access_token';
const REFRESH_KEY = 'admin_template_refresh_token';

export function getAccessToken(): string {
  return localStorage.getItem(ACCESS_KEY) ?? '';
}

export function setAccessToken(token: string): void {
  if (token) localStorage.setItem(ACCESS_KEY, token);
  else localStorage.removeItem(ACCESS_KEY);
}

export function getRefreshToken(): string {
  return localStorage.getItem(REFRESH_KEY) ?? '';
}

export function setRefreshToken(token: string): void {
  if (token) localStorage.setItem(REFRESH_KEY, token);
  else localStorage.removeItem(REFRESH_KEY);
}

export function clearTokens(): void {
  localStorage.removeItem(ACCESS_KEY);
  localStorage.removeItem(REFRESH_KEY);
}
