import { http } from './request'

export const orderApi = {
	// Student
	list: (params) => http.get('/orders', params),
	detail: (id) => http.get('/orders/' + id),
	pay: (id, data) => http.post('/orders/' + id + '/pay', data || {}),
	checkin: (id, data) => http.post('/orders/' + id + '/checkin', data),
	complete: (id) => http.post('/orders/' + id + '/complete'),
	cancel: (id, data) => http.post('/orders/' + id + '/cancel', data),
	review: (id, data) => http.post('/orders/' + id + '/review', data),

	// Employer
	employerList: (params) => http.get('/employer/orders', params),
	employerDetail: (id) => http.get('/employer/orders/' + id),
	confirm: (id) => http.post('/employer/orders/' + id + '/confirm'),
	employerCancel: (id, data) => http.post('/employer/orders/' + id + '/cancel', data),
	employerReview: (id, data) => http.post('/employer/orders/' + id + '/review', data),

	// Admin
	adminList: (params) => http.get('/admin/orders', params),
}
