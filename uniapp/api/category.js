import { http } from './request'

export const categoryApi = {
	list: (status) => http.get('/categories', status ? { status } : {}, { auth: false }),
	// Admin
	adminList: () => http.get('/admin/categories'),
	create: (data) => http.post('/admin/categories', data),
	update: (id, data) => http.put('/admin/categories/' + id, data),
	delete: (id) => http.delete('/admin/categories/' + id),
}
