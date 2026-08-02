<template>
	<view class="page apps-page">
		<view class="tabs">
			<view v-for="t in tabs" :key="t.value" class="tab" :class="{ active: filter === t.value }" @click="setFilter(t.value)">
				<text>{{ t.label }}</text>
			</view>
		</view>

		<view class="list">
			<view v-for="a in list" :key="a.id" class="card app" @click="goDetail(a)">
				<view class="row1">
					<text class="title">{{ a.job_title }}</text>
					<text class="tag" :class="statusTag(a.status)">{{ statusText(a.status) }}</text>
				</view>
				<view class="row2" v-if="role==='employer'">
					<text class="muted">报名人:{{ a.student_name || '匿名' }} · {{ a.student_school || '—' }}</text>
				</view>
				<view class="row2" v-else>
					<text class="muted">联系电话:{{ a.contact_phone || '—' }}</text>
				</view>
				<view class="row3">
					<text class="muted">{{ formatTime(a.created_at) }}</text>
				</view>

				<!-- 雇主审核按钮 -->
				<view v-if="role==='employer' && a.status===1" class="actions">
					<button class="btn-secondary" @click.stop="audit(a, 3)">拒绝</button>
					<button class="btn-primary" @click.stop="audit(a, 2)">通过</button>
				</view>
				<!-- 学生取消 -->
				<view v-if="role==='student' && a.status===1" class="actions">
					<button class="btn-ghost" @click.stop="cancel(a)">取消报名</button>
				</view>
			</view>
			<empty-state v-if="!loading && list.length === 0" emoji="📋" text="还没有任何报名" />
		</view>
	</view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { jobApi } from '@/api/job'
	import EmptyState from '@/components/empty-state/empty-state.vue'
	import { APP_STATUS, APP_STATUS_TEXT, APP_STATUS_TAG } from '@/utils/constants'
	import { formatDateTime } from '@/utils/format'
	import { confirm, toastSuccess, toastError } from '@/utils/ui'

	export default {
		components: { EmptyState },
		data() {
			return {
				filter: 0,
				list: [],
				loading: false,
				tabs: [
					{ value: 0, label: '全部' },
					{ value: 1, label: '待审核' },
					{ value: 2, label: '已通过' },
					{ value: 5, label: '已录用' },
					{ value: 3, label: '已拒绝' },
					{ value: 4, label: '已取消' },
				],
			}
		},
		computed: {
			role() {
				const u = useUserStore()
				return u.activeRole
			},
		},
		onShow() { this.refresh() },
		onPullDownRefresh() {
			this.refresh().finally(() => uni.stopPullDownRefresh())
		},
		methods: {
			setFilter(v) { this.filter = v; this.refresh() },
			statusText(s) { return APP_STATUS_TEXT[s] || '—' },
			statusTag(s) { return APP_STATUS_TAG[s === 5 ? 2 : s] || 'tag-info' },
			formatTime: formatDateTime,
			goDetail(a) {
				uni.navigateTo({ url: '/pages/application-detail/application-detail?id=' + a.id })
			},
			async refresh() {
				this.loading = true
				try {
					const params = { page: 1, size: 50 }
					if (this.filter) params.status = this.filter
					if (this.role === 'student') {
						const res = await jobApi.myApplications(params)
						this.list = (res && res.list) || []
					} else if (this.role === 'employer') {
						// Employer view: pull from each of their jobs.
						const myJobsRes = await jobApi.myJobs({ page: 1, size: 100 })
						const jobs = (myJobsRes && myJobsRes.list) || []
						const all = []
						for (const j of jobs) {
							const r = await jobApi.jobApplications(j.id, { page: 1, size: 100 })
							const arr = (r && r.list) || []
							all.push(...arr)
						}
						this.list = this.filter ? all.filter(a => a.status === this.filter) : all
					}
				} catch (e) {} finally {
					this.loading = false
				}
			},
			async audit(a, action) {
				const isPass = action === 2
				const ok = await confirm(isPass ? '通过该报名?' : '拒绝该报名?')
				if (!ok) return
				try {
					await jobApi.auditApplication(a.id, { action, remark: isPass ? '欢迎加入' : '暂不匹配' })
					toastSuccess(isPass ? '已通过' : '已拒绝')
					this.refresh()
				} catch (e) { toastError('操作失败') }
			},
			async cancel(a) {
				const ok = await confirm('确认取消报名?')
				if (!ok) return
				try {
					await jobApi.cancelApplication(a.id)
					toastSuccess('已取消')
					this.refresh()
				} catch (e) { toastError('操作失败') }
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.tabs { display: flex; gap: $spacing-xs; padding: $spacing-sm $spacing-sm; overflow-x: auto; white-space: nowrap; }
	.tab {
		padding: 12rpx 24rpx; background: $bg-card; border-radius: $radius-pill;
		font-size: $font-sm; color: $text-regular; flex-shrink: 0;
		&.active { background: $brand-primary; color: #fff; }
	}
	.list { padding: 0 $spacing-sm; }
	.app { margin-bottom: $spacing-sm; }
	.row1 { display: flex; align-items: center; justify-content: space-between; margin-bottom: $spacing-xs; }
	.title { font-size: $font-lg; font-weight: 600; color: $text-primary; flex: 1; }
	.row2, .row3 { font-size: $font-sm; color: $text-secondary; margin-top: 4rpx; }
	.muted { color: $text-secondary; }
	.actions { display: flex; gap: $spacing-sm; margin-top: $spacing-md; justify-content: flex-end; }
	.btn-primary, .btn-secondary, .btn-ghost { font-size: $font-sm; padding: 12rpx 32rpx; }
</style>
