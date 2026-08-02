// API base URL — read from storage at runtime so the user can switch
// between dev / staging / LAN-IP without rebuilding the mini-program.
//
// Why a module-level mutable + getter instead of a const? The mini-program
// runs in a single JS context; the URL is read once on every request
// (cheap) and writes are infrequent (settings page). Avoiding reactivity
// here keeps the request layer decoupled from any store.

const DEFAULT_URL = 'http://localhost:8080'
const STORAGE_KEY = 'campus_gig_api_base_url'

let cached = ''

function read() {
	if (cached) return cached
	try {
		const v = uni.getStorageSync(STORAGE_KEY)
		cached = (v && typeof v === 'string') ? v : DEFAULT_URL
	} catch (e) {
		cached = DEFAULT_URL
	}
	return cached
}

function write(url) {
	cached = url || DEFAULT_URL
	try { uni.setStorageSync(STORAGE_KEY, cached) } catch (e) {}
}

export function getApiBase() {
	return read()
}

export function getFullApiBase() {
	return read() + '/api/v1'
}

export function setApiBase(url) {
	write(url)
}

export function resetApiBase() {
	cached = ''
	try { uni.removeStorageSync(STORAGE_KEY) } catch (e) {}
}

// Helper for the settings page — returns the current effective value along
// with a small "is custom" hint so the UI can show "已自定义" vs "默认".
export function getApiBaseMeta() {
	const v = read()
	return {
		url: v,
		isCustom: v !== DEFAULT_URL,
		default: DEFAULT_URL,
	}
}
