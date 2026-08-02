// Lightweight toast helpers. We don't use uni.showToast directly in pages so
// the visual style is consistent across the app.

export function toast(title, icon = 'none', duration = 1800) {
	uni.showToast({ title: String(title || ''), icon, duration })
}

export function toastSuccess(title, duration) {
	return toast(title, 'success', duration)
}

export function toastError(title, duration) {
	return toast(title, 'error', duration)
}

export function toastLoading(title = '加载中') {
	uni.showLoading({ title, mask: true })
}

export function hideLoading() {
	uni.hideLoading()
}

export function confirm(content, title = '提示') {
	return new Promise(resolve => {
		uni.showModal({
			title, content, showCancel: true,
			confirmText: '确定', cancelText: '取消',
			success: r => resolve(r.confirm),
			fail: () => resolve(false),
		})
	})
}
