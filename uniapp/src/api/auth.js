import { http } from './request'

export const authApi = {
	login: (data) => http.post('/auth/login', data, { auth: false }),
	register: (data) => http.post('/auth/register', data, { auth: false }),
	logout: (data) => http.post('/auth/logout', data || {}),
	me: () => http.get('/auth/me'),
	refresh: (refreshToken) => http.post('/auth/refresh', { refresh_token: refreshToken }, { auth: false }),
}
