import { http } from './request'

export const jobApi = {
	// Public
	list: (params) => http.get('/jobs', params, { auth: false }),
	get: (id) => http.get('/jobs/' + id, {}, { auth: false }),
	apply: (id, data) => http.post('/jobs/' + id + '/apply', data),

	// Student-side
	myApplications: (params) => http.get('/applications', params),
	cancelApplication: (id) => http.post('/applications/' + id + '/cancel'),
	applicationDetail: (id) => http.get('/applications/' + id),

	// Employer-side
	myJobs: (params) => http.get('/employer/jobs', params),
	create: (data) => http.post('/employer/jobs', data),
	update: (id, data) => http.put('/employer/jobs/' + id, data),
	offline: (id) => http.post('/employer/jobs/' + id + '/offline'),
	delete: (id) => http.delete('/employer/jobs/' + id),
	jobApplications: (id, params) => http.get('/employer/jobs/' + id + '/applications', params),
	auditApplication: (id, data) => http.post('/employer/applications/' + id + '/audit', data),
	hire: (id, data) => http.post('/employer/applications/' + id + '/hire', data),

	// Admin
	pending: (params) => http.get('/admin/jobs', params),
	audit: (id, data) => http.post('/admin/jobs/' + id + '/audit', data),
}
