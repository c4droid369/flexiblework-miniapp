import { http } from './request'

export const messageApi = {
	list: (params) => http.get('/messages', params),
	markRead: (id) => http.post('/messages/' + id + '/read'),
	markAllRead: () => http.post('/messages/read-all'),
	// Admin
	broadcast: (data) => http.post('/admin/messages/broadcast', data),
}
