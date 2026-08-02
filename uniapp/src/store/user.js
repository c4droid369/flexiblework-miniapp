import { defineStore } from 'pinia'
import { authApi } from '@/api/auth'
import {
	getToken, setToken, getRefreshToken, setRefreshToken,
	clearTokens, getStoredUser, setStoredUser, clearStoredUser,
	getActiveRole, setActiveRole,
} from '@/utils/auth'
import { USER_TYPES } from '@/utils/constants'

// User store. Owns the access/refresh tokens, the cached /me payload, and
// the active role the tabbar renders for. `bootstrap` is called on app
// launch to restore state from local storage; `login` / `register` set it
// fresh from the server.
export const useUserStore = defineStore('user', {
	state: () => ({
		token: '',
		refreshToken: '',
		user: null,        // /me payload (MeResp)
		activeRole: '',    // admin | student | employer
		loading: false,
	}),
	getters: {
		isLoggedIn: (s) => !!s.token && !!s.user,
		userType: (s) => (s.user && s.user.user_type) || '',
		userTypes: (s) => (s.user && s.user.user_types) || [],
		permissions: (s) => (s.user && s.user.permissions) || [],
		isStudent: (s) => s.activeRole === USER_TYPES.STUDENT,
		isEmployer: (s) => s.activeRole === USER_TYPES.EMPLOYER,
		can: (s) => (code) => s.permissions.includes(code),
	},
	actions: {
		bootstrap() {
			this.token = getToken()
			this.refreshToken = getRefreshToken()
			const u = getStoredUser()
			if (u) this.user = u
			const r = getActiveRole()
			// Prefer server-declared user_type over stale storage, fallback to
			// last-active role.
			this.activeRole = (u && u.user_type) || r || ''
		},
		async login(payload) {
			this.loading = true
			try {
				const data = await authApi.login(payload)
				this.token = data.access_token
				this.refreshToken = data.refresh_token
				setToken(data.access_token)
				setRefreshToken(data.refresh_token)
				await this.fetchMe()
				return data
			} finally {
				this.loading = false
			}
		},
		async register(payload) {
			this.loading = true
			try {
				const data = await authApi.register(payload)
				this.token = data.access_token
				this.refreshToken = data.refresh_token
				setToken(data.access_token)
				setRefreshToken(data.refresh_token)
				await this.fetchMe()
				return data
			} finally {
				this.loading = false
			}
		},
		async fetchMe() {
			const me = await authApi.me()
			this.user = me
			this.activeRole = me.user_type
			setStoredUser(me)
			setActiveRole(me.user_type)
			return me
		},
		async logout() {
			try { await authApi.logout({ refresh_token: this.refreshToken }) } catch (e) {}
			this.token = ''
			this.refreshToken = ''
			this.user = null
			this.activeRole = ''
			clearTokens()
			clearStoredUser()
		},
		// Switch active role for a user with multiple roles (e.g. an admin
		// who is also a student). Persisted so the tabbar is stable across
		// cold starts.
		switchRole(role) {
			if (!this.user || !this.user.user_types.includes(role)) return
			this.activeRole = role
			setActiveRole(role)
		},
	},
})
