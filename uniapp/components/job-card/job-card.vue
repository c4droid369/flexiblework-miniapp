<template>
	<view class="job-card card" @click="onClick">
		<view class="row1">
			<text class="title">{{ job.title }}</text>
			<text class="salary">{{ salaryText }}</text>
		</view>
		<view class="row2" v-if="job.employer_name || job.category_name">
			<text class="muted" v-if="job.category_name">{{ job.category_name }}</text>
			<text class="muted" v-if="job.employer_name"> · {{ job.employer_name }}</text>
		</view>
		<view class="row3">
			<text class="muted">📍 {{ job.location || '不限' }}</text>
		</view>
		<view class="row4" v-if="extra">
			<slot name="extra"></slot>
		</view>
		<view class="tags" v-if="tags.length">
			<text v-for="(t, i) in tags" :key="i" class="tag" :class="t.cls">{{ t.text }}</text>
		</view>
	</view>
</template>

<script>
	import { JOB_STATUS_TEXT, JOB_STATUS_TAG } from '@/utils/constants'
	import { buildSalaryText } from '@/utils/format'

	export default {
		name: 'JobCard',
		props: {
			job: { type: Object, required: true },
			extra: { type: Boolean, default: false },
		},
		computed: {
			salaryText() { return buildSalaryText(this.job) },
		tags() {
			const t = []
			if (this.job.status != null && this.job.status !== 2) {
				t.push({
					text: JOB_STATUS_TEXT[this.job.status] || '—',
					cls: JOB_STATUS_TAG[this.job.status] || 'tag-info',
				})
			}
			return t
		},
		},
		methods: {
			onClick() {
				this.$emit('click', this.job)
				uni.navigateTo({ url: '/pages/job-detail/job-detail?id=' + this.job.id })
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.job-card { margin-bottom: $spacing-sm; }
	.row1 { display: flex; align-items: center; justify-content: space-between; margin-bottom: $spacing-xs; }
	.title { font-size: $font-lg; font-weight: 600; color: $text-primary; flex: 1; }
	.salary { font-size: $font-lg; color: $brand-primary; font-weight: 600; margin-left: $spacing-sm; }
	.row2, .row3 { font-size: $font-sm; margin-top: 4rpx; }
	.muted { color: $text-secondary; }
	.tags { display: flex; gap: $spacing-xs; margin-top: $spacing-sm; flex-wrap: wrap; }
</style>
