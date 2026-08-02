// API base URL is **not** a static constant here. See `@/utils/api-base.js` —
// it reads the URL from storage at runtime so the user can switch dev /
// staging / LAN-IP targets without rebuilding.
//
// `localhost` 在小程序里指的是**小程序进程自己**,不是宿主机:
//   - 微信开发者工具(自带模拟器): http://localhost:8080   ✅
//   - iOS 模拟器(Mac):            http://localhost:8080   ✅
//   - Android 模拟器:              http://10.0.2.2:8080    ← 不要写 localhost
//   - 真机(扫码预览):             http://<电脑局域网IP>:8080
//
// 真机调试时先 `ipconfig` 查局域网 IP,再在小程序"我的 → 服务器设置"里改。

// Storage keys (other than API URL, which lives in utils/api-base.js).
export const STORAGE_KEYS = {
	ACCESS_TOKEN: 'campus_gig_access_token',
	REFRESH_TOKEN: 'campus_gig_refresh_token',
	USER_INFO: 'campus_gig_user_info',
	ACTIVE_ROLE: 'campus_gig_active_role',
}

// User type vocabulary. Mirrors backend.
export const USER_TYPES = {
	ADMIN: 'admin',
	STUDENT: 'student',
	EMPLOYER: 'employer',
}

// Job status.
export const JOB_STATUS = {
	DRAFT: 0, PENDING: 1, RECRUITING: 2, OFFLINE: 3,
	REJECTED: 4, FILLED: 5,
}

export const JOB_STATUS_TEXT = {
	0: '草稿', 1: '待审核', 2: '招聘中', 3: '已下架',
	4: '审核未通过', 5: '已招满',
}

// 岗位状态色板。和其它 *_TAG 一一对应。
//   0 草稿 / 3 已下架  → info (中性灰)
//   1 待审核          → warning (黄,提醒)
//   2 招聘中          → success (绿,主流程)
//   4 审核未通过      → danger (红,异常)
//   5 已招满          → primary (橙,流程结束态)
export const JOB_STATUS_TAG = {
	0: 'tag-info',
	1: 'tag-warning',
	2: 'tag-success',
	3: 'tag-info',
	4: 'tag-danger',
	5: 'tag-primary',
}

// Application status.
export const APP_STATUS = {
	PENDING: 1, APPROVED: 2, REJECTED: 3, CANCELLED: 4, HIRED: 5,
}

export const APP_STATUS_TEXT = {
	1: '待审核', 2: '已通过', 3: '已拒绝', 4: '已取消', 5: '已录用',
}

// App status pill colors. status 5 (已录用) 视觉上等同 2 (已通过),
// 所以 page 里 `APP_STATUS_TAG[s === 5 ? 2 : s]` 直接拿这个表就行。
export const APP_STATUS_TAG = {
	1: 'tag-warning',  // 待审核
	2: 'tag-success',  // 已通过
	3: 'tag-danger',   // 已拒绝
	4: 'tag-info',     // 已取消
	5: 'tag-success',  // 已录用
}

// Order status.
export const ORDER_STATUS = {
	WAIT_PAY: 1, PAID: 2, IN_PROGRESS: 3, WAIT_CONFIRM: 4,
	SETTLED: 5, CANCELLED: 6, REFUNDED: 7,
}

export const ORDER_STATUS_TEXT = {
	1: '待支付', 2: '已支付', 3: '进行中', 4: '待确认完成',
	5: '已结算', 6: '已取消', 7: '已退款',
}

export const ORDER_STATUS_TAG = {
	1: 'tag-warning', 2: 'tag-info', 3: 'tag-primary',
	4: 'tag-warning', 5: 'tag-success', 6: 'tag-info', 7: 'tag-info',
}

// Cert status.
export const CERT_STATUS = {
	NONE: 0, PENDING: 1, APPROVED: 2, REJECTED: 3,
}

export const CERT_STATUS_TEXT = {
	0: '未认证', 1: '审核中', 2: '已认证', 3: '已拒绝',
}

export const CERT_STATUS_TAG = {
	0: 'tag-info', 1: 'tag-warning', 2: 'tag-success', 3: 'tag-danger',
}

// Salary type.
export const SALARY_TYPE = {
	HOURLY: 1, DAILY: 2, WEEKLY: 3, MONTHLY: 4, PER_TASK: 5,
}
export const SALARY_TYPE_TEXT = { 1: '时薪', 2: '日薪', 3: '周薪', 4: '月薪', 5: '按件' }

// Settlement type.
export const SETTLE_TYPE_TEXT = { 1: '日结', 2: '周结', 3: '完工结' }
