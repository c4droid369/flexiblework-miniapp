import { http } from './request'

export const profileApi = {
	// Student
	getStudent: () => http.get('/student/profile'),
	updateStudent: (data) => http.post('/student/profile', data),
	submitStudentCert: (data) => http.post('/student/certification', data),

	// Employer
	getEmployer: () => http.get('/employer/profile'),
	updateEmployer: (data) => http.post('/employer/profile', data),
	submitEmployerCert: (data) => http.post('/employer/certification', data),

	// Campus agent (校园代理) — same business flow as employer, separate profile.
	getAgent: () => http.get('/agent/profile'),
	updateAgent: (data) => http.post('/agent/profile', data),
	submitAgentCert: (data) => http.post('/agent/certification', data),

	// Admin cert review
	listPendingStudentCerts: () => http.get('/admin/student-certifications'),
	auditStudentCert: (id, data) => http.post('/admin/student-certifications/' + id + '/audit', data),
	listPendingEmployerCerts: () => http.get('/admin/employer-certifications'),
	auditEmployerCert: (id, data) => http.post('/admin/employer-certifications/' + id + '/audit', data),
	listPendingAgentCerts: () => http.get('/admin/agent-certifications'),
	auditAgentCert: (id, data) => http.post('/admin/agent-certifications/' + id + '/audit', data),
}
