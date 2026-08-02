<template>
	<view class="page job-detail" v-if="job">
		<view class="hero card">
			<view class="title-row">
				<text class="title">{{ job.title }}</text>
				<text class="salary">{{ job.salary_text || formatSalary }}</text>
			</view>
			<view class="meta">
				<text class="cat" v-if="job.category_name">📂 {{ job.category_name }}</text>
				<text class="emp" v-if="job.employer_name">🏪 {{ job.employer_name }}</text>
			</view>
			<view class="info-row">
				<text class="ico">📍</text>
				<text class="t">{{ job.location || '不限' }}</text>
			</view>
			<view class="info-row" v-if="job.work_time_start">
				<text class="ico">⏰</text>
				<text class="t">{{ job.work_time_start }} - {{ job.work_time_end }}</text>
			</view>
			<view class="info-row" v-if="job.recruit_count">
				<text class="ico">👥</text>
				<text class="t">招 {{ job.recruit_count }} 人 · {{ settleText }} · {{ salaryTypeText }}</text>
			</view>
		</view>

		<view class="section card">
			<view class="section-title">岗位描述</view>
			<text class="desc">{{ job.description }}</text>
		</view>

		<view class="section card" v-if="job.requirements">
			<view class="section-title">岗位要求</view>
			<text class="desc">{{ job.requirements }}</text>
		</view>

		<view class="section card" v-if="job.employer_name">
			<view class="section-title">发布方</view>
			<view class="emp-row">
				<text class="emp-name">{{ job.employer_name }}</text>
				<text class="muted">已认证雇主</text>
			</view>
		</view>

		<!-- 底部操作栏 -->
		<view class="footer safe-bottom" v-if="canShowApply">
			<button class="btn-secondary" v-if="isEmployerMine" @click="offline">下架</button>
			<button class="btn-primary flex-1" v-if="canApply" @click="onApply">立即报名</button>
			<view v-else-if="alreadyApplied" class="applied-tag tag tag-success">已报名</view>
		</view>
	</view>
	<view v-else class="page"><empty-state text="加载中..." emoji="⏳" /></view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { jobApi } from '@/api/job'
	import EmptyState from '@/components/empty-state/empty-state.vue'
	import { SALARY_TYPE_TEXT, SETTLE_TYPE_TEXT } from '@/utils/constants'
	import { buildSalaryText } from '@/utils/format'
	import { confirm, toastSuccess, toastError } from '@/utils/ui'

	export default {
		components: { EmptyState },
		data() {
			return {
				job: null,
				alreadyApplied: false,
			}
		},
		computed: {
			user() { return useUserStore() },
			salaryTypeText() { return SALARY_TYPE_TEXT[this.job.salary_type] || '' },
			settleText() { return SETTLE_TYPE_TEXT[this.job.settlement_type] || '' },
			formatSalary() { return buildSalaryText(this.job) },
			canApply() {
				return this.user.activeRole === 'student' && this.job.status === 2 && !this.alreadyApplied
			},
			isEmployerMine() {
				return this.user.activeRole === 'employer' && this.user.user && this.job && this.user.user.id === this.job.employer_id
			},
			canShowApply() {
				return this.canApply || this.isEmployerMine || this.alreadyApplied
			},
		},
		onLoad(query) {
			this.id = query.id
			this.load()
		},
		methods: {
			async load() {
				try {
					this.job = await jobApi.get(this.id)
				} catch (e) { return }
				if (this.user.activeRole === 'student') {
					try {
						const res = await jobApi.myApplications({ page: 1, size: 100 })
						const apps = (res && res.list) || []
						this.alreadyApplied = apps.some(a => a.job_id === this.job.id && a.status !== 4)
					} catch (e) {}
				}
			},
			async onApply() {
				if (this.user.user.cert_status === undefined) {
					// not yet loaded; fall through to optimistic submit
				}
				// The backend will return Forbidden if the student is not
				// certified — surface that as a friendly prompt.
				try {
					await jobApi.apply(this.id, { message: '想报名这个岗位', contact_phone: this.user.user.phone || '' })
					toastSuccess('报名成功,等待审核')
					this.alreadyApplied = true
				} catch (e) {
					const msg = e.message || ''
					if (msg.includes('认证') || msg.includes('实名')) {
						const ok = await confirm(msg + '\n是否前往认证?')
						if (ok) uni.navigateTo({ url: '/pages/certification/certification' })
					}
				}
			},
			async offline() {
				const ok = await confirm('确认下架该岗位?')
				if (!ok) return
				try {
					await jobApi.offline(this.id)
					toastSuccess('已下架')
					setTimeout(() => uni.navigateBack(), 600)
				} catch (e) { toastError('下架失败') }
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.page { padding-bottom: 180rpx; }
	.hero { margin: $spacing-sm $spacing-sm 0; }
	.title-row { display: flex; align-items: flex-start; justify-content: space-between; }
	.title { font-size: 40rpx; font-weight: 600; color: $text-primary; flex: 1; line-height: 1.3; }
	.salary { font-size: 36rpx; color: $brand-primary; font-weight: 600; margin-left: $spacing-sm; }
	.meta { display: flex; gap: $spacing-sm; margin-top: $spacing-sm; flex-wrap: wrap; }
	.cat, .emp { background: $brand-primary-bg; color: $brand-primary; padding: 6rpx 16rpx; border-radius: $radius-pill; font-size: $font-xs; }
	.info-row { display: flex; align-items: center; margin-top: $spacing-sm; }
	.info-row .ico { margin-right: $spacing-sm; font-size: $font-md; }
	.info-row .t { color: $text-regular; font-size: $font-base; }
	.section { margin: $spacing-sm; }
	.section-title { font-size: $font-md; font-weight: 600; color: $text-primary; margin-bottom: $spacing-sm; }
	.desc { font-size: $font-base; color: $text-regular; line-height: $line-relaxed; white-space: pre-wrap; }
	.emp-row { display: flex; align-items: center; gap: $spacing-sm; }
	.emp-name { font-size: $font-base; color: $text-primary; font-weight: 500; }
	.muted { color: $text-secondary; font-size: $font-sm; }
	.footer {
		position: fixed; left: 0; right: 0; bottom: 0; padding: $spacing-sm $spacing-md;
		background: $bg-card; display: flex; gap: $spacing-sm; align-items: center;
		box-shadow: 0 -2rpx 16rpx rgba(0,0,0,0.06);
	}
	.applied-tag { flex: 1; text-align: center; padding: 20rpx 0; }
	.flex-1 { flex: 1; }
</style>
