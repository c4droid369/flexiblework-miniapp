<template>
	<view class="page my-jobs-page">
		<view class="list">
			<view v-for="j in list" :key="j.id" class="card job" @click="goDetail(j)">
				<view class="row1">
					<text class="title">{{ j.title }}</text>
					<text class="tag" :class="statusTag(j.status)">{{ statusText(j.status) }}</text>
				</view>
				<view class="row2">
					<text class="muted">{{ j.salary_text }}</text>
					<text class="muted"> · 招 {{ j.recruit_count }} 人</text>
				</view>
				<view class="row3">
					<text class="muted">📍 {{ j.location || '不限' }}</text>
				</view>
				<view class="row4">
					<text class="muted">浏览 {{ j.view_count }} · 报名 {{ j.apply_count }}</text>
				</view>
				<view class="actions" v-if="j.status===2">
					<button class="btn-secondary" @click.stop="offline(j)">下架</button>
				</view>
			</view>
			<empty-state v-if="!loading && list.length===0" emoji="📋" text="还没有发布过岗位" action-text="去发布" @action="goPublish" />
		</view>

		<view class="floating">
			<button class="btn-primary" @click="goPublish">+ 发布岗位</button>
		</view>
	</view>
</template>

<script>
	import { jobApi } from '@/api/job'
	import EmptyState from '@/components/empty-state/empty-state.vue'
	import { JOB_STATUS_TEXT, JOB_STATUS_TAG } from '@/utils/constants'
	import { confirm, toastSuccess, toastError } from '@/utils/ui'

	export default {
		components: { EmptyState },
		data() { return { list: [], loading: false } },
		onShow() { this.refresh() },
		onPullDownRefresh() { this.refresh().finally(() => uni.stopPullDownRefresh()) },
		methods: {
			statusText: (s) => JOB_STATUS_TEXT[s] || '—',
			statusTag: (s) => JOB_STATUS_TAG[s] || 'tag-info',
			goPublish() { uni.navigateTo({ url: '/pages/job-publish/job-publish' }) },
			goDetail(j) { uni.navigateTo({ url: '/pages/job-detail/job-detail?id=' + j.id }) },
			async refresh() {
				this.loading = true
				try {
					const r = await jobApi.myJobs({ page: 1, size: 100 })
					this.list = (r && r.list) || []
				} catch (e) {} finally { this.loading = false }
			},
			async offline(j) {
				const ok = await confirm('确认下架该岗位?')
				if (!ok) return
				try {
					await jobApi.offline(j.id)
					toastSuccess('已下架')
					this.refresh()
				} catch (e) { toastError('下架失败') }
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.list { padding: $spacing-sm; }
	.job { margin-bottom: $spacing-sm; }
	.row1 { display: flex; align-items: center; justify-content: space-between; }
	.title { font-size: $font-lg; font-weight: 600; color: $text-primary; flex: 1; }
	.row2, .row3, .row4 { font-size: $font-sm; color: $text-secondary; margin-top: 4rpx; }
	.muted { color: $text-secondary; }
	.actions { display: flex; justify-content: flex-end; margin-top: $spacing-sm; }
	.btn-secondary { font-size: $font-sm; padding: 8rpx 24rpx; }
	.floating {
		position: fixed; right: $spacing-md; bottom: $spacing-md;
	}
</style>
