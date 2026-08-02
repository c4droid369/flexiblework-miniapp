// Format helpers — keep page templates terse by pre-formatting on the model.

/** Format an ISO timestamp or Date as `YYYY-MM-DD HH:mm` (local time). */
export function formatDateTime(t) {
	if (!t) return ''
	const d = t instanceof Date ? t : new Date(t)
	if (isNaN(d.getTime())) return ''
	const pad = n => (n < 10 ? '0' + n : '' + n)
	return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
		' ' + pad(d.getHours()) + ':' + pad(d.getMinutes())
}

/** Date only. */
export function formatDate(t) {
	if (!t) return ''
	const d = t instanceof Date ? t : new Date(t)
	if (isNaN(d.getTime())) return ''
	const pad = n => (n < 10 ? '0' + n : '' + n)
	return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate())
}

/** Relative time, e.g. "刚刚 / 5 分钟前 / 3 小时前 / 昨天 / 2025-08-02". */
export function timeAgo(t) {
	if (!t) return ''
	const d = t instanceof Date ? t : new Date(t)
	if (isNaN(d.getTime())) return ''
	const diff = Math.floor((Date.now() - d.getTime()) / 1000)
	if (diff < 60) return '刚刚'
	if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前'
	if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前'
	if (diff < 86400 * 2) return '昨天'
	return formatDate(t)
}

/** Format RMB amount — `150.5` → `"¥150.50"`. */
export function formatMoney(n) {
	if (n == null || n === '') return ''
	const num = Number(n)
	if (isNaN(num)) return ''
	return '¥' + num.toFixed(2)
}

/** Strip the leading "周" or empty parts out of a work_time string. */
export function formatWorkTime(start, end) {
	if (!start && !end) return '时间灵活'
	if (start && end) return start + ' - ' + end
	return start || end || ''
}

/** Safe string fallback. */
export function s(v, fallback = '') {
	if (v == null) return fallback
	return String(v)
}

/** Build a salary text from server-provided fields when the server didn't. */
export function buildSalaryText(job) {
	if (!job) return ''
	if (job.salary_text) return job.salary_text
	const min = Number(job.salary_min || 0)
	const max = Number(job.salary_max || 0)
	const unit = job.salary_unit || '元'
	if (max > 0 && min !== max) return min + '-' + max + unit
	return min + unit
}
