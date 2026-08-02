// API base URL — two layers of configurability:
//
//   1. **Build-time** (VITE_API_BASE_URL env var). When you build the
//      mini-program, set this to your deployment's API origin and it
//      becomes the first-run default — no manual setup required for
//      end users.
//
//        $ VITE_API_BASE_URL=https://api.example.com:8082 \
//          npm run build:h5
//
//   2. **Runtime** (settings page). The "我的 → 服务器设置" page lets the
//      end user override the value at any time and persists it to
//      local storage. Useful when the same build artifact is reused
//      across multiple deployments (e.g. dev + staging) and the user
//      wants to flip between them without re-downloading.
//
// Why a module-level mutable + getter instead of a const? The mini-program
// runs in a single JS context; the URL is read once on every request
// (cheap) and writes are infrequent (settings page). Avoiding reactivity
// here keeps the request layer decoupled from any store.

const BUILD_URL = (import.meta.env && import.meta.env.VITE_API_BASE_URL) || ''
const DEFAULT_URL = BUILD_URL || 'http://localhost:8080'
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
