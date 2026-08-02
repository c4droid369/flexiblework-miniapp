// Local storage helpers. Wraps uni.getStorageSync / setStorageSync with
// JSON serialization and silent failure (storage may be unavailable on some
// platforms; we treat that as "no data" rather than throwing).

import { STORAGE_KEYS } from './constants'

export function getToken() {
	try {
		return uni.getStorageSync(STORAGE_KEYS.ACCESS_TOKEN) || ''
	} catch (e) {
		return ''
	}
}

export function setToken(t) {
	try { uni.setStorageSync(STORAGE_KEYS.ACCESS_TOKEN, t) } catch (e) {}
}

export function getRefreshToken() {
	try {
		return uni.getStorageSync(STORAGE_KEYS.REFRESH_TOKEN) || ''
	} catch (e) {
		return ''
	}
}

export function setRefreshToken(t) {
	try { uni.setStorageSync(STORAGE_KEYS.REFRESH_TOKEN, t) } catch (e) {}
}

export function clearTokens() {
	try {
		uni.removeStorageSync(STORAGE_KEYS.ACCESS_TOKEN)
		uni.removeStorageSync(STORAGE_KEYS.REFRESH_TOKEN)
	} catch (e) {}
}

export function getStoredUser() {
	try {
		const raw = uni.getStorageSync(STORAGE_KEYS.USER_INFO)
		return raw ? JSON.parse(raw) : null
	} catch (e) {
		return null
	}
}

export function setStoredUser(u) {
	try { uni.setStorageSync(STORAGE_KEYS.USER_INFO, JSON.stringify(u)) } catch (e) {}
}

export function clearStoredUser() {
	try { uni.removeStorageSync(STORAGE_KEYS.USER_INFO) } catch (e) {}
}

export function getActiveRole() {
	try {
		return uni.getStorageSync(STORAGE_KEYS.ACTIVE_ROLE) || ''
	} catch (e) {
		return ''
	}
}

export function setActiveRole(role) {
	try { uni.setStorageSync(STORAGE_KEYS.ACTIVE_ROLE, role) } catch (e) {}
}
