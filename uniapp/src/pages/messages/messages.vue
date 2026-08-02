<template>
	<view class="page messages-page">
		<view class="list">
			<view v-for="m in list" :key="m.id" class="card msg" :class="{ unread: !m.is_read }" @click="onClick(m)">
				<view class="row1">
					<text class="type-tag tag" :class="typeTag(m.type)">{{ typeText(m.type) }}</text>
					<text class="title">{{ m.title }}</text>
				</view>
				<text class="content">{{ m.content }}</text>
				<text class="time">{{ timeAgo(m.created_at) }}</text>
			</view>
			<empty-state v-if="!loading && list.length===0" emoji="🔔" text="暂无消息" />
		</view>

		<view class="floating" v-if="hasUnread">
			<button class="btn-secondary" @click="markAll">全部标为已读</button>
		</view>
	</view>
</template>

<script>
	import { messageApi } from '@/api/message'
	import EmptyState from '@/components/empty-state/empty-state.vue'
	import { timeAgo } from '@/utils/format'
	import { toastSuccess } from '@/utils/ui'

	export default {
		components: { EmptyState },
		data() { return { list: [], loading: false } },
		computed: {
			hasUnread() { return this.list.some(m => !m.is_read) },
		},
		onShow() { this.refresh() },
		onPullDownRefresh() { this.refresh().finally(() => uni.stopPullDownRefresh()) },
		methods: {
			typeText: (t) => ({ 1: '系统', 2: '岗位', 3: '订单', 4: '评价' }[t] || '消息'),
			typeTag: (t) => ({ 1: 'tag-info', 2: 'tag-primary', 3: 'tag-warning', 4: 'tag-success' }[t] || 'tag-info'),
			timeAgo,
			async refresh() {
				this.loading = true
				try {
					const r = await messageApi.list({ page: 1, size: 50 })
					this.list = (r && r.list) || []
				} catch (e) {} finally { this.loading = false }
			},
			async onClick(m) {
				if (!m.is_read) {
					try { await messageApi.markRead(m.id) } catch (e) {}
					m.is_read = true
				}
				if (m.link) uni.navigateTo({ url: m.link, fail: () => {} })
			},
			async markAll() {
				try { await messageApi.markAllRead() } catch (e) {}
				toastSuccess('已全部标为已读')
				this.refresh()
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.list { padding: $spacing-sm; }
	.msg { margin-bottom: $spacing-sm; position: relative; }
	.msg.unread::before {
		content: ''; position: absolute; left: 16rpx; top: 32rpx; width: 16rpx; height: 16rpx;
		background: $brand-primary; border-radius: 50%;
	}
	.msg.unread { padding-left: 40rpx; }
	.row1 { display: flex; align-items: center; gap: $spacing-xs; margin-bottom: 6rpx; }
	.title { font-size: $font-md; font-weight: 600; color: $text-primary; }
	.content { display: block; font-size: $font-base; color: $text-regular; line-height: $line-relaxed; margin: 4rpx 0; }
	.time { font-size: $font-xs; color: $text-placeholder; }
	.floating { position: fixed; right: $spacing-md; bottom: $spacing-md; }
</style>
