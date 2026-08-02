<template>
	<view class="page orders-page">
		<view class="tabs">
			<view v-for="t in tabs" :key="t.value" class="tab" :class="{ active: filter === t.value }" @click="setFilter(t.value)">
				<text>{{ t.label }}</text>
			</view>
		</view>

		<view class="list">
			<view v-for="o in list" :key="o.id" class="card order" @click="goDetail(o)">
				<view class="row1">
					<text class="title">{{ o.job_title }}</text>
					<text class="tag" :class="statusTag(o.status)">{{ statusText(o.status) }}</text>
				</view>
				<view class="row2">
					<text class="muted">订单号: {{ o.order_no }}</text>
				</view>
				<view class="row3">
					<text class="amount">{{ formatMoney(o.amount) }}</text>
					<text class="muted">{{ otherParty(o) }}</text>
				</view>
				<view class="row4">
					<text class="muted">{{ formatTime(o.created_at) }}</text>
				</view>
			</view>
			<empty-state v-if="!loading && list.length === 0" emoji="📦" text="暂无订单" />
		</view>
	</view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { orderApi } from '@/api/order'
	import EmptyState from '@/components/empty-state/empty-state.vue'
	import { ORDER_STATUS, ORDER_STATUS_TEXT, ORDER_STATUS_TAG } from '@/utils/constants'
	import { formatMoney, formatDateTime } from '@/utils/format'

	export default {
		components: { EmptyState },
		data() {
			return {
				filter: 0,
				list: [],
				loading: false,
				tabs: [
					{ value: 0, label: '全部' },
					{ value: 1, label: '待支付' },
					{ value: 2, label: '已支付' },
					{ value: 3, label: '进行中' },
					{ value: 4, label: '待确认' },
					{ value: 5, label: '已结算' },
				],
			}
		},
		computed: {
			role() { return useUserStore().activeRole },
		},
		onShow() { this.refresh() },
		onPullDownRefresh() {
			this.refresh().finally(() => uni.stopPullDownRefresh())
		},
		methods: {
			setFilter(v) { this.filter = v; this.refresh() },
			statusText: (s) => ORDER_STATUS_TEXT[s] || '—',
			statusTag: (s) => ORDER_STATUS_TAG[s] || 'tag-info',
			formatMoney, formatTime: formatDateTime,
			otherParty(o) {
				return this.role === 'student' ? '雇主:' + (o.employer_name || '—') : '学生:' + (o.student_name || '—')
			},
			goDetail(o) { uni.navigateTo({ url: '/pages/order-detail/order-detail?id=' + o.id }) },
			async refresh() {
				this.loading = true
				try {
					const params = { page: 1, size: 50 }
					if (this.filter) params.status = this.filter
					const fn = this.role === 'student' ? orderApi.list : orderApi.employerList
					const res = await fn(params)
					this.list = (res && res.list) || []
				} catch (e) {} finally {
					this.loading = false
				}
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
	.order { margin-bottom: $spacing-sm; }
	.row1 { display: flex; align-items: center; justify-content: space-between; margin-bottom: $spacing-xs; }
	.title { font-size: $font-lg; font-weight: 600; color: $text-primary; flex: 1; }
	.row2, .row3, .row4 { font-size: $font-sm; color: $text-secondary; margin-top: 4rpx; }
	.muted { color: $text-secondary; }
	.amount { color: $brand-primary; font-size: $font-lg; font-weight: 600; margin-right: $spacing-md; }
</style>
