<template>
	<!-- Role-aware home. Renders student jobs view or employer dashboard. -->
	<view class="page">
		<view v-if="!user.isLoggedIn" class="not-logged">
			<view class="hero">
				<view class="logo">🎓</view>
				<view class="title">校园灵活用工</view>
				<view class="subtitle">靠谱兼职 · 一站搞定</view>
			</view>
			<view class="actions">
				<button class="btn-primary" @click="goLogin">立即登录</button>
				<button class="btn-secondary" @click="goRegister">注册账号</button>
			</view>
			<view class="demo">
				<view class="title">体验账号(点击直接填充)</view>
				<view class="acc" @click="goFill('demo_student', 'demo123')">
					<text class="role">学生</text>
					<text>demo_student / demo123</text>
				</view>
				<view class="acc" @click="goFill('demo_employer', 'demo123')">
					<text class="role">雇主</text>
					<text>demo_employer / demo123</text>
				</view>
			</view>
			<view class="server-link" @click="goSettings">⚙️ 服务器设置(连不上后端点这里)</view>
		</view>

		<view v-else-if="user.activeRole === 'student'" class="student-home">
			<view class="header">
				<view class="greet">
					<text class="hi">Hi, {{ nickname }}</text>
					<text class="sub">想找点什么兼职?</text>
				</view>
				<view class="msg-btn" @click="goMessages">
					<text class="ico">🔔</text>
				</view>
			</view>

			<view class="search" @click="onSearch">
				<text class="ico">🔍</text>
				<text class="placeholder">搜索岗位 / 地点 / 关键字</text>
			</view>

			<scroll-view scroll-x class="cats" :show-scrollbar="false">
				<view class="cat" :class="{ active: !currentCat }" @click="pickCat(null)">
					<text>全部</text>
				</view>
				<view v-for="c in categories" :key="c.id" class="cat" :class="{ active: currentCat === c.id }" @click="pickCat(c.id)">
					<text>{{ c.name }}</text>
				</view>
			</scroll-view>

			<view class="list">
				<job-card v-for="j in list" :key="j.id" :job="j" />
				<empty-state v-if="!loading && list.length === 0" emoji="🍃" text="没有匹配的岗位" />
				<view v-if="loading" class="loading"><text>加载中...</text></view>
			</view>
		</view>

		<view v-else-if="user.activeRole === 'employer'" class="emp-home">
			<view class="header emp">
				<view class="greet">
					<text class="hi">{{ nickname }}, 下午好</text>
					<text class="sub">今天有 {{ stats.pending_apps }} 份报名待审核</text>
				</view>
				<view class="msg-btn" @click="goMessages">
					<text class="ico">🔔</text>
				</view>
			</view>

			<view class="grid">
				<view class="grid-item" @click="goMyJobs()">
					<text class="ico">📋</text>
					<text class="num">{{ stats.total_jobs }}</text>
					<text class="t">在招岗位</text>
				</view>
				<view class="grid-item" @click="goMyJobs(1)">
					<text class="ico">⏳</text>
					<text class="num">{{ stats.pending_apps }}</text>
					<text class="t">待审核报名</text>
				</view>
				<view class="grid-item" @click="goOrders(3)">
					<text class="ico">📦</text>
					<text class="num">{{ stats.in_progress }}</text>
					<text class="t">进行中订单</text>
				</view>
				<view class="grid-item" @click="goOrders(5)">
					<text class="ico">💰</text>
					<text class="num">{{ stats.completed_orders }}</text>
					<text class="t">已完成订单</text>
				</view>
			</view>

			<view class="quick card">
				<view class="qi primary" @click="goPublish">
					<text class="ico">➕</text>
					<view>
						<text class="t">发布新岗位</text>
						<text class="d">几秒钟发布,招到合适的学生</text>
					</view>
				</view>
				<view class="qi" @click="goMyJobs()">
					<text class="ico">📋</text>
					<view>
						<text class="t">岗位管理</text>
						<text class="d">查看 / 上下架 / 编辑</text>
					</view>
				</view>
				<view class="qi" @click="goApps">
					<text class="ico">📝</text>
					<view>
						<text class="t">收到的报名</text>
						<text class="d">审核通过 / 拒绝 / 录用</text>
					</view>
				</view>
			</view>
		</view>

		<view v-else class="not-logged">
			<view class="hero">
				<view class="logo">⚙️</view>
				<view class="title">管理员请前往管理后台</view>
				<view class="subtitle">本小程序仅服务学生/雇主</view>
			</view>
			<button class="btn-primary" @click="onLogout">退出登录</button>
		</view>
	</view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { useAppStore } from '@/store/app'
	import { jobApi } from '@/api/job'
	import { orderApi } from '@/api/order'
	import { profileApi } from '@/api/profile'
	import JobCard from '@/components/job-card/job-card.vue'
	import EmptyState from '@/components/empty-state/empty-state.vue'
	import { toastSuccess } from '@/utils/ui'

	export default {
		components: { JobCard, EmptyState },
		data() {
			return {
				keyword: '',
				currentCat: null,
				list: [],
				loading: false,
				stats: { total_jobs: 0, pending_apps: 0, in_progress: 0, completed_orders: 0 },
			}
		},
		computed: {
			user() { return useUserStore() },
			nickname() {
				const u = this.user
				return (u.user && (u.user.nickname || u.user.username)) || '同学'
			},
			categories() {
				const app = useAppStore()
				return app.activeCategories
			},
		},
		onShow() {
			const app = useAppStore()
			app.loadCategories()
			if (this.user.activeRole === 'student') this.refresh()
			if (this.user.activeRole === 'employer') this.loadEmpStats()
		},
		onPullDownRefresh() {
			const done = () => uni.stopPullDownRefresh()
			if (this.user.activeRole === 'student') this.refresh().finally(done)
			else if (this.user.activeRole === 'employer') this.loadEmpStats().finally(done)
			else done()
		},
		methods: {
			goLogin() { uni.navigateTo({ url: '/pages/auth/login' }) },
			goRegister() { uni.redirectTo({ url: '/pages/auth/role-select' }) },
			goSettings() { uni.navigateTo({ url: '/pages/settings/settings' }) },
			goFill() {
				// Direct user to login page; one-tap fill is on the login page itself.
				uni.navigateTo({ url: '/pages/auth/login' })
			},
			goMessages() { uni.navigateTo({ url: '/pages/messages/messages' }) },
			goPublish() { uni.navigateTo({ url: '/pages/job-publish/job-publish' }) },
			goMyJobs() { uni.navigateTo({ url: '/pages/my-jobs/my-jobs' }) },
			goOrders(status) { uni.switchTab({ url: '/pages/orders/orders?status=' + (status || '') }) },
			goApps() { uni.switchTab({ url: '/pages/applications/applications' }) },
			async onLogout() {
				await this.user.logout()
				toastSuccess('已退出')
				setTimeout(() => uni.reLaunch({ url: '/pages/auth/role-select' }), 400)
			},
			onSearch() {
				uni.showModal({
					title: '搜索',
					editable: true,
					placeholderText: '输入关键字',
					success: r => {
						if (r.confirm && r.content) {
							this.keyword = r.content
							this.refresh()
						}
					},
				})
			},
			pickCat(id) { this.currentCat = id; this.refresh() },
			async refresh() {
				this.loading = true
				try {
					const params = { page: 1, size: 10 }
					if (this.currentCat) params.category_id = this.currentCat
					if (this.keyword) params.keyword = this.keyword
					const res = await jobApi.list(params)
					this.list = (res && res.list) || []
				} catch (e) {} finally {
					this.loading = false
				}
			},
			async loadEmpStats() {
				try {
					const p = await profileApi.getEmployer()
					this.stats.total_jobs = (p && p.total_jobs) || 0
					this.stats.completed_orders = (p && p.completed_orders) || 0
				} catch (e) {}
				try {
					const my = await jobApi.myJobs({ page: 1, size: 100 })
					const jobs = (my && my.list) || []
					let pending = 0
					for (const j of jobs) {
						const r = await jobApi.jobApplications(j.id, { page: 1, size: 100 })
						pending += ((r && r.list) || []).filter(a => a.status === 1).length
					}
					this.stats.pending_apps = pending
				} catch (e) {}
				try {
					const o = await orderApi.employerList({ page: 1, size: 100, status: 3 })
					this.stats.in_progress = (o && o.total) || 0
				} catch (e) {}
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.page { min-height: 100vh; background: $bg-page; }
	.not-logged { padding: 100rpx $spacing-md; text-align: center; }
	.hero { margin-bottom: $spacing-xl; }
	.logo { font-size: 160rpx; }
	.title { font-size: 44rpx; font-weight: 600; color: $text-primary; margin-top: 12rpx; }
	.subtitle { font-size: $font-base; color: $text-secondary; margin-top: 12rpx; }
	.actions { display: flex; flex-direction: column; gap: $spacing-sm; }
	.btn-primary, .btn-secondary { width: 100%; }
	.demo {
		margin-top: $spacing-lg; background: $brand-primary-bg; border-radius: $radius-md;
		padding: $spacing-md; text-align: left;
		.title { font-size: $font-sm; color: $text-secondary; margin-bottom: $spacing-sm; }
		.acc { display: flex; align-items: center; gap: $spacing-sm; background: $bg-card; border-radius: $radius-sm; padding: 12rpx $spacing-sm; margin-bottom: $spacing-xs; }
		.role { background: $brand-primary; color: #fff; font-size: $font-xs; padding: 2rpx 12rpx; border-radius: $radius-pill; }
	}
	.server-link { margin-top: $spacing-lg; text-align: center; color: $brand-primary; font-size: $font-sm; padding: $spacing-sm; }

	.student-home .header {
		background: linear-gradient(135deg, $brand-primary 0%, $brand-primary-dark 100%);
		color: #fff; padding: 40rpx $spacing-md $spacing-lg; border-bottom-left-radius: $radius-xl; border-bottom-right-radius: $radius-xl;
		display: flex; align-items: center; justify-content: space-between;
		.greet { flex: 1; }
		.hi { font-size: 36rpx; font-weight: 600; }
		.sub { display: block; font-size: $font-sm; opacity: 0.9; margin-top: 6rpx; }
		.msg-btn { width: 80rpx; height: 80rpx; display: flex; align-items: center; justify-content: center; background: rgba(255,255,255,0.2); border-radius: 50%; }
		.ico { font-size: 36rpx; }
	}
	.search {
		background: $bg-card; margin: -$spacing-md $spacing-md 0; padding: 24rpx $spacing-md;
		border-radius: $radius-pill; display: flex; align-items: center; gap: $spacing-sm;
		box-shadow: $shadow-sm; position: relative; z-index: 2;
		.ico { color: $text-secondary; }
		.placeholder { color: $text-placeholder; font-size: $font-base; }
	}
	.cats { white-space: nowrap; padding: $spacing-md $spacing-sm $spacing-sm; }
	.cat {
		display: inline-block; padding: 12rpx 28rpx; margin-right: $spacing-sm;
		background: $bg-card; border-radius: $radius-pill;
		font-size: $font-sm; color: $text-regular;
		&.active { background: $brand-primary; color: #fff; }
	}
	.list { padding: 0 $spacing-sm $spacing-md; }
	.loading { text-align: center; padding: $spacing-md; color: $text-placeholder; font-size: $font-sm; }

	.emp-home .header.emp {
		background: linear-gradient(135deg, $brand-secondary 0%, #2EA85A 100%);
		color: #fff; padding: 40rpx $spacing-md $spacing-lg; border-bottom-left-radius: $radius-xl; border-bottom-right-radius: $radius-xl;
		display: flex; align-items: center; justify-content: space-between;
		.greet { flex: 1; }
		.hi { font-size: 36rpx; font-weight: 600; }
		.sub { display: block; font-size: $font-sm; opacity: 0.9; margin-top: 6rpx; }
		.msg-btn { width: 80rpx; height: 80rpx; display: flex; align-items: center; justify-content: center; background: rgba(255,255,255,0.2); border-radius: 50%; }
		.ico { font-size: 36rpx; }
	}
	.emp-home .grid {
		display: grid; grid-template-columns: repeat(4, 1fr); gap: $spacing-sm;
		background: $bg-card; margin: -$spacing-md $spacing-sm 0; padding: $spacing-md $spacing-xs;
		border-radius: $radius-md; box-shadow: $shadow-sm; position: relative; z-index: 2;
		.grid-item { display: flex; flex-direction: column; align-items: center; padding: $spacing-sm 0; }
		.ico { font-size: 48rpx; }
		.num { font-size: 36rpx; font-weight: 600; color: $text-primary; margin-top: 4rpx; }
		.t { font-size: $font-xs; color: $text-secondary; margin-top: 4rpx; }
	}
	.emp-home .quick { margin: $spacing-md $spacing-sm; }
	.qi {
		display: flex; align-items: center; padding: $spacing-md;
		border-bottom: 2rpx solid $border-color-light; gap: $spacing-md;
		&:last-child { border-bottom: none; }
		.ico { font-size: 48rpx; }
		.t { display: block; font-size: $font-base; font-weight: 600; color: $text-primary; }
		.d { display: block; font-size: $font-sm; color: $text-secondary; margin-top: 4rpx; }
		&.primary { background: $brand-primary-bg; }
		&.primary .ico { color: $brand-primary; }
	}
</style>
