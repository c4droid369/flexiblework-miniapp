// Network layer — wraps uni.request with auth header injection, automatic
// refresh-on-401, and a uniform envelope unwrap. The frontend sees only the
// `data` payload; HTTP errors throw a typed ApiError for catch blocks.

import { getFullApiBase } from '@/utils/api-base'
import {
	getToken, setToken,
	getRefreshToken, setRefreshToken,
	clearTokens, clearStoredUser,
} from '@/utils/auth'

export class ApiError extends Error {
	constructor(code, message, httpStatus) {
		super(message)
		this.code = code
		this.httpStatus = httpStatus
	}
}

// In-flight refresh dedup — multiple 401s share one refresh request.
let refreshInflight = null

async function doRefresh() {
	const rt = getRefreshToken()
	if (!rt) throw new ApiError(401, '登录已过期,请重新登录', 401)
	if (refreshInflight) return refreshInflight
	refreshInflight = uni.request({
		url: getFullApiBase() + '/auth/refresh',
		method: 'POST',
		header: { 'Content-Type': 'application/json' },
		data: { refresh_token: rt },
	}).then(res => {
		refreshInflight = null
		const body = typeof res.data === 'string' ? JSON.parse(res.data) : res.data
		if (res.statusCode !== 200 || body.code !== 0) {
			throw new ApiError(body.code || 500, body.message || 'refresh failed', res.statusCode)
		}
		setToken(body.data.access_token)
		setRefreshToken(body.data.refresh_token)
		return body.data.access_token
	}).catch(err => {
		refreshInflight = null
		throw err
	})
	return refreshInflight
}

// Clear session and bounce to login.
function forceLogout() {
	clearTokens()
	clearStoredUser()
	uni.showToast({ title: '登录已过期', icon: 'none' })
	setTimeout(() => {
		uni.reLaunch({ url: '/pages/auth/role-select' })
	}, 800)
}

export function request({
	url, method = 'GET', data, header = {}, silent = false, auth = true,
}) {
	return new Promise((resolve, reject) => {
		const headers = { ...header }
		if (auth) {
			const t = getToken()
			if (t) headers['Authorization'] = 'Bearer ' + t
		}
		if (method !== 'GET' && data !== undefined && !headers['Content-Type']) {
			headers['Content-Type'] = 'application/json'
		}

		uni.request({
			url: url.startsWith('http') ? url : getFullApiBase() + url,
			method,
			data,
			header: headers,
			timeout: 15000,
			success: async (res) => {
				const body = typeof res.data === 'string' ? JSON.parse(res.data || '{}') : res.data
				// Backend returns HTTP 200 even for application errors but with
				// code != 0. Treat anything else as success.
				if (res.statusCode === 200 && body && body.code === 0) {
					return resolve(body.data)
				}
				// 401 → try refresh once.
				if (res.statusCode === 401 && auth) {
					try {
						await doRefresh()
						return resolve(request({ url, method, data, header, silent, auth }))
					} catch (e) {
						forceLogout()
						return reject(e)
					}
				}
				const err = new ApiError(
					body && body.code != null ? body.code : res.statusCode,
					(body && body.message) || ('http ' + res.statusCode),
					res.statusCode,
				)
				if (!silent) uni.showToast({ title: err.message, icon: 'none' })
				reject(err)
			},
			fail: (err) => {
				const e = new ApiError(-1, err.errMsg || '网络错误', 0)
				if (!silent) uni.showToast({ title: e.message, icon: 'none' })
				reject(e)
			},
		})
	})
}

export const http = {
	get: (url, data, opts) => request({ url, method: 'GET', data, ...(opts || {}) }),
	post: (url, data, opts) => request({ url, method: 'POST', data, ...(opts || {}) }),
	put: (url, data, opts) => request({ url, method: 'PUT', data, ...(opts || {}) }),
	delete: (url, data, opts) => request({ url, method: 'DELETE', data, ...(opts || {}) }),
}
