<template>
	<view class="page mine-page">
		<view class="header" :class="headerClass">
			<view class="row">
				<view class="avatar">{{ avatarChar }}</view>
				<view class="info">
					<text class="name">{{ nickname }}</text>
					<text class="role">{{ roleText }}</text>
				</view>
				<view v-if="canSwitch" class="switch-btn" @click="onSwitchRole">
					<text>切换身份</text>
				</view>
			</view>
			<!-- 雇主 / 校园代理:核心数据 -->
			<view v-if="role==='employer' || role==='agent'" class="stats">
				<view class="stat">
					<text class="num">{{ stats.total_jobs }}</text>
					<text class="lbl">在招岗位</text>
				</view>
				<view class="stat">
					<text class="num">{{ stats.completed_orders }}</text>
					<text class="lbl">完成订单</text>
				</view>
				<view class="stat">
					<text class="num">{{ stats.rating.toFixed(1) }}</text>
					<text class="lbl">评分</text>
				</view>
			</view>
			<!-- 学生:认证 banner -->
			<view v-else class="cert-banner" @click="goCert">
				<text class="lbl">学生认证</text>
				<text class="val" :class="'tag-' + certTag">{{ certText }}</text>
				<text class="arrow">›</text>
			</view>
			<!-- 校园代理:额外显示推荐码(可分享) -->
			<view v-if="role==='agent' && referralCode" class="refcode">
				<text class="ref-lbl">我的推荐码</text>
				<text class="ref-val" @click="copyReferral">{{ referralCode }}</text>
			</view>
		</view>

		<view class="quick">
			<view class="quick-item" @click="goMyJobs" v-if="role==='employer' || role==='agent'">
				<text class="ico">📋</text>
				<text class="t">我的岗位</text>
			</view>
			<view class="quick-item" @click="goPublish" v-if="role==='employer' || role==='agent'">
				<text class="ico">➕</text>
				<text class="t">发布岗位</text>
			</view>
			<view class="quick-item" @click="goProfile" v-if="role==='employer' || role==='agent'">
				<text class="ico">🤝</text>
				<text class="t">{{ role==='agent' ? '代理资料' : '店铺资料' }}</text>
			</view>
			<view class="quick-item" @click="goCert" v-if="role==='student'">
				<text class="ico">🎓</text>
				<text class="t">学生认证</text>
			</view>
			<view class="quick-item" @click="goCert" v-if="role==='agent'">
				<text class="ico">✅</text>
				<text class="t">资质认证</text>
			</view>
			<view class="quick-item" @click="goMessages">
				<text class="ico">🔔</text>
				<text class="t">消息中心</text>
			</view>
		</view>

		<view class="menu card">
			<view class="menu-row" @click="goCert">
				<text class="lbl">{{ role==='agent' ? '校园代理资质' : '实名认证' }}</text>
				<text class="val">{{ certText }}</text>
			</view>
			<view class="menu-row" @click="goProfile">
				<text class="lbl">{{ profileLbl }}</text>
				<text class="arrow">›</text>
			</view>
			<view class="menu-row" @click="goMessages">
				<text class="lbl">消息中心</text>
				<text class="arrow">›</text>
			</view>
			<view class="menu-row" @click="goSettings">
				<text class="lbl">服务器设置</text>
				<text class="arrow">›</text>
			</view>
			<view class="menu-row" @click="onLogout">
				<text class="lbl" style="color:#F5222D">退出登录</text>
				<text class="arrow">›</text>
			</view>
		</view>

		<view class="footer muted">
			<text>校园灵活用工小程序 · v1.0</text>
		</view>
	</view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { profileApi } from '@/api/profile'
	import { CERT_STATUS_TEXT, CERT_STATUS_TAG } from '@/utils/constants'
	import { confirm, toastSuccess, toast } from '@/utils/ui'

	export default {
		data() {
			return {
				stats: { total_jobs: 0, completed_orders: 0, rating: 5.0 },
				certText: '未认证',
				certTag: 'info',
				referralCode: '',
			}
		},
		computed: {
			role() { return useUserStore().activeRole },
			nickname() {
				const u = useUserStore()
				return (u.user && (u.user.nickname || u.user.username)) || '未登录'
			},
			avatarChar() { return (this.nickname || '?').slice(0, 1).toUpperCase() },
			headerClass() {
				if (this.role === 'employer') return 'is-employer'
				if (this.role === 'agent') return 'is-agent'
				return ''
			},
			roleText() {
				const map = { student: '学生', employer: '雇主', agent: '校园代理', admin: '管理员' }
				return map[this.role] || '游客'
			},
			profileLbl() {
				if (this.role === 'employer') return '店铺资料'
				if (this.role === 'agent') return '代理资料'
				return '个人资料'
			},
			canSwitch() {
				const u = useUserStore()
				return u.userTypes && u.userTypes.length > 1
			},
		},
		onShow() {
			if (this.role === 'employer') this.loadEmployer()
			if (this.role === 'student') this.loadStudent()
			if (this.role === 'agent') this.loadAgent()
		},
		methods: {
			async loadEmployer() {
				try {
					const p = await profileApi.getEmployer()
					if (p) {
						this.certText = CERT_STATUS_TEXT[p.cert_status] || '未认证'
						this.certTag = CERT_STATUS_TAG[p.cert_status] || 'info'
						this.stats.total_jobs = p.total_jobs || 0
						this.stats.completed_orders = p.completed_orders || 0
						this.stats.rating = p.rating || 5.0
					}
				} catch (e) {}
			},
			async loadStudent() {
				try {
					const p = await profileApi.getStudent()
					if (p) {
						this.certText = CERT_STATUS_TEXT[p.cert_status] || '未认证'
						this.certTag = CERT_STATUS_TAG[p.cert_status] || 'info'
					}
				} catch (e) {}
			},
			async loadAgent() {
				try {
					const p = await profileApi.getAgent()
					if (p) {
						this.certText = CERT_STATUS_TEXT[p.cert_status] || '未认证'
						this.certTag = CERT_STATUS_TAG[p.cert_status] || 'info'
						this.stats.total_jobs = p.total_jobs || 0
						this.stats.completed_orders = 0 // agent 暂无 completed_orders 字段
						this.stats.rating = p.rating || 5.0
						this.referralCode = p.referral_code || ''
					}
				} catch (e) {}
			},
			copyReferral() {
				if (!this.referralCode) return
				uni.setClipboardData({ data: this.referralCode })
				toast('已复制推荐码')
			},
			goCert() { uni.navigateTo({ url: '/pages/certification/certification' }) },
			goProfile() { uni.navigateTo({ url: '/pages/certification/certification?tab=profile' }) },
			goMessages() { uni.navigateTo({ url: '/pages/messages/messages' }) },
			goSettings() { uni.navigateTo({ url: '/pages/settings/settings' }) },
			goMyJobs() { uni.navigateTo({ url: '/pages/my-jobs/my-jobs' }) },
			goPublish() { uni.navigateTo({ url: '/pages/job-publish/job-publish' }) },
			async onLogout() {
				const ok = await confirm('确定退出登录?')
				if (!ok) return
				const u = useUserStore()
				await u.logout()
				toastSuccess('已退出')
				setTimeout(() => uni.reLaunch({ url: '/pages/auth/role-select' }), 500)
			},
			onSwitchRole() {
				const u = useUserStore()
				const list = (u.userTypes || []).filter(r => r !== 'admin')
				if (list.length < 2) return
				const labelMap = { student: '学生', employer: '雇主', agent: '校园代理' }
				uni.showActionSheet({
					itemList: list.map(r => labelMap[r] || r),
					success: r => {
						u.switchRole(list[r.tapIndex])
						uni.showToast({ title: '已切换', icon: 'success' })
						setTimeout(() => uni.reLaunch({ url: '/pages/index/index' }), 500)
					},
				})
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.mine-page { padding-top: 0; }
	.header {
		background: linear-gradient(135deg, $brand-primary 0%, $brand-primary-dark 100%);
		color: #fff; padding: 40rpx $spacing-md $spacing-lg;
		border-bottom-left-radius: $radius-xl; border-bottom-right-radius: $radius-xl;
		&.is-employer { background: linear-gradient(135deg, $brand-secondary 0%, #2EA85A 100%); }
		&.is-agent { background: linear-gradient(135deg, $brand-accent 0%, #E89B1E 100%); }
		.row { display: flex; align-items: center; }
		.avatar { width: 100rpx; height: 100rpx; border-radius: 50%; background: rgba(255,255,255,0.25); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 44rpx; font-weight: 600; }
		.info { flex: 1; margin-left: $spacing-md; }
		.name { font-size: 36rpx; font-weight: 600; display: block; }
		.role { display: inline-block; font-size: $font-xs; background: rgba(255,255,255,0.25); padding: 4rpx 16rpx; border-radius: $radius-pill; margin-top: 8rpx; }
		.switch-btn { background: rgba(255,255,255,0.25); padding: 10rpx 20rpx; border-radius: $radius-pill; font-size: $font-sm; }
		.stats { display: flex; justify-content: space-around; margin-top: $spacing-md; }
		.stat { text-align: center; }
		.num { display: block; font-size: 36rpx; font-weight: 600; }
		.lbl { font-size: $font-xs; opacity: 0.9; }
		.cert-banner { background: rgba(255,255,255,0.2); border-radius: $radius-md; padding: 20rpx $spacing-md; margin-top: $spacing-md; display: flex; align-items: center; }
		.cert-banner .lbl { font-size: $font-base; }
		.cert-banner .val { flex: 1; text-align: right; }
		.cert-banner .arrow { font-size: 32rpx; opacity: 0.7; margin-left: 8rpx; }
		.refcode {
			background: rgba(255,255,255,0.22); border-radius: $radius-md;
			padding: 16rpx $spacing-md; margin-top: $spacing-sm;
			display: flex; align-items: center; justify-content: space-between;
			.ref-lbl { font-size: $font-sm; opacity: 0.9; }
			.ref-val { font-size: $font-lg; font-weight: 600; font-family: monospace; letter-spacing: 2rpx; }
		}
	}
	.quick {
		display: grid; grid-template-columns: repeat(4, 1fr); gap: $spacing-sm;
		background: $bg-card; margin: -$spacing-md $spacing-sm 0; padding: $spacing-md $spacing-sm;
		border-radius: $radius-md; box-shadow: $shadow-sm; position: relative; z-index: 2;
		.quick-item { display: flex; flex-direction: column; align-items: center; padding: $spacing-sm 0; }
		.ico { font-size: 48rpx; }
		.t { font-size: $font-sm; color: $text-regular; margin-top: 8rpx; }
	}
	.menu { margin: $spacing-md $spacing-sm; }
	.menu-row {
		display: flex; align-items: center; padding: 24rpx $spacing-sm;
		border-bottom: 2rpx solid $border-color-light;
		&:last-child { border-bottom: none; }
		.lbl { flex: 1; font-size: $font-base; color: $text-primary; }
		.val { font-size: $font-sm; color: $text-secondary; }
		.arrow { color: $text-placeholder; font-size: 32rpx; }
	}
	.footer { text-align: center; padding: $spacing-md; font-size: $font-xs; }
	.muted { color: $text-placeholder; }
</style>