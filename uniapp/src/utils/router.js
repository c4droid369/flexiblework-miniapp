// Lightweight router helpers — thin wrappers over uni.navigateTo / redirectTo /
// switchTab that auto-prefix the tabbar check.

export function go(url, replace = false) {
	if (replace) return uni.redirectTo({ url })
	return uni.navigateTo({ url })
}

export function goTab(url) {
	return uni.switchTab({ url })
}

export function back(delta = 1, fallback = '/pages/index/index') {
	const pages = getCurrentPages()
	if (pages.length > delta) {
		uni.navigateBack({ delta })
	} else {
		uni.switchTab({ url: fallback })
	}
}

export function relaunch(url) {
	return uni.reLaunch({ url })
}
