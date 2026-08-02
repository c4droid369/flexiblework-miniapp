<template>
	<view class="page dashboard-page">
		<view class="hero card">
			<view class="row">
				<view class="avatar">🏪</view>
				<view class="info">
					<text class="t">欢迎回来,{{ nickname }}</text>
					<text class="d">今天是 {{ today }},也是招聘的好日子</text>
				</view>
			</view>
		</view>

		<view class="grid card">
			<view class="cell" @click="go('/pages/my-jobs/my-jobs')">
				<text class="num">{{ stats.total_jobs }}</text>
				<text class="t">在招岗位</text>
			</view>
			<view class="cell" @click="goApps">
				<text class="num">{{ stats.pending_apps }}</text>
				<text class="t">待审核报名</text>
			</view>
			<view class="cell" @click="goOrders(3)">
				<text class="num">{{ stats.in_progress }}</text>
				<text class="t">进行中订单</text>
			</view>
			<view class="cell" @click="goOrders(5)">
				<text class="num">{{ stats.completed_orders }}</text>
				<text class="t">已完成订单</text>
			</view>
		</view>

		<view class="quick card">
			<view class="qi primary" @click="go('/pages/job-publish/job-publish')">
				<text class="ico">➕</text>
				<view><text class="t">发布新岗位</text><text class="d">几秒钟发布,招到合适的学生</text></view>
			</view>
			<view class="qi" @click="go('/pages/my-jobs/my-jobs')">
				<text class="ico">📋</text>
				<view><text class="t">岗位管理</text><text class="d">查看 / 上下架 / 编辑</text></view>
			</view>
			<view class="qi" @click="goApps">
				<text class="ico">📝</text>
				<view><text class="t">收到的报名</text><text class="d">审核通过 / 拒绝 / 录用</text></view>
			</view>
			<view class="qi" @click="go('/pages/certification/certification')">
				<text class="ico">🏪</text>
				<view><text class="t">店铺资料</text><text class="d">维护公司信息、认证状态</text></view>
			</view>
		</view>
	</view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { jobApi } from '@/api/job'
	import { orderApi } from '@/api/order'
	import { profileApi } from '@/api/profile'

	export default {
		data() {
			return {
				stats: { total_jobs: 0, pending_apps: 0, in_progress: 0, completed_orders: 0 },
			}
		},
		computed: {
			user() { return useUserStore() },
			nickname() { return (this.user.user && (this.user.user.nickname || this.user.user.username)) || '老板' },
			today() {
				const d = new Date()
				return d.getMonth() + 1 + '月' + d.getDate() + '日'
			},
		},
		onShow() { this.load() },
		methods: {
			go(url) { uni.navigateTo({ url }) },
			goApps() { uni.switchTab({ url: '/pages/applications/applications' }) },
			goOrders(status) { uni.switchTab({ url: '/pages/orders/orders?status=' + (status || '') }) },
			async load() {
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
	.hero {
		background: linear-gradient(135deg, $brand-secondary 0%, #2EA85A 100%);
		color: #fff; padding: $spacing-md; margin: $spacing-sm;
		.row { display: flex; align-items: center; }
		.avatar { font-size: 64rpx; }
		.info { flex: 1; margin-left: $spacing-md; }
		.t { display: block; font-size: 36rpx; font-weight: 600; }
		.d { display: block; font-size: $font-sm; opacity: 0.9; margin-top: 4rpx; }
	}
	.grid { display: grid; grid-template-columns: repeat(4, 1fr); margin: 0 $spacing-sm $spacing-sm; }
	.cell { text-align: center; padding: $spacing-sm 0; }
	.num { display: block; font-size: 40rpx; font-weight: 600; color: $text-primary; }
	.t { display: block; font-size: $font-xs; color: $text-secondary; margin-top: 4rpx; }
	.quick { margin: 0 $spacing-sm; }
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
