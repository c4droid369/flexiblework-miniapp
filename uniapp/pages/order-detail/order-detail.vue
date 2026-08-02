<template>
	<view class="page order-detail" v-if="o">
		<!-- 顶部状态条 -->
		<view class="status-bar" :class="'st-' + o.status">
			<text class="st-text">{{ statusText }}</text>
			<text class="st-sub" v-if="statusSub">{{ statusSub }}</text>
		</view>

		<view class="card hero">
			<text class="title">{{ o.job_title }}</text>
			<view class="amount-row">
				<text class="amount">{{ formatMoney(o.amount) }}</text>
				<text class="muted">订单号 {{ o.order_no }}</text>
			</view>
			<view class="row" v-if="role==='student'"><text class="lbl">雇主</text><text class="val">{{ o.employer_name }}</text></view>
			<view class="row" v-else><text class="lbl">学生</text><text class="val">{{ o.student_name }}</text></view>
			<view class="row"><text class="lbl">下单时间</text><text class="val">{{ formatTime(o.created_at) }}</text></view>
			<view class="row" v-if="o.paid_at"><text class="lbl">支付时间</text><text class="val">{{ formatTime(o.paid_at) }}</text></view>
			<view class="row" v-if="o.started_at"><text class="lbl">开始时间</text><text class="val">{{ formatTime(o.started_at) }}</text></view>
			<view class="row" v-if="o.completed_at"><text class="lbl">完成时间</text><text class="val">{{ formatTime(o.completed_at) }}</text></view>
			<view class="row" v-if="o.settled_at"><text class="lbl">结算时间</text><text class="val">{{ formatTime(o.settled_at) }}</text></view>
			<view class="row" v-if="o.cancel_reason"><text class="lbl">取消原因</text><text class="val">{{ o.cancel_reason }}</text></view>
		</view>

		<view class="card" v-if="o.work_proof && o.work_proof.length">
			<view class="section-title">上岗凭证</view>
			<view class="proof-grid">
				<image v-for="(p, i) in o.work_proof" :key="i" :src="p" class="proof-img" mode="aspectFill" @click="preview(i)" />
			</view>
		</view>

		<!-- 状态机操作 -->
		<view class="footer safe-bottom">
			<template v-if="role==='student' && o.status===1">
				<button class="btn-ghost" @click="onCancel">取消订单</button>
				<button class="btn-primary flex-1" @click="onPay">立即支付</button>
			</template>
			<template v-if="role==='student' && o.status===2">
				<button class="btn-primary flex-1" @click="onCheckin">上岗打卡</button>
			</template>
			<template v-if="role==='student' && o.status===3">
				<button class="btn-primary flex-1" @click="onComplete">提交完成</button>
			</template>
			<template v-if="role==='student' && o.status===5 && !reviewed">
				<button class="btn-primary flex-1" @click="onReview">评价雇主</button>
			</template>
			<template v-if="role==='employer' && o.status===4">
				<button class="btn-ghost" @click="onCancel">取消订单</button>
				<button class="btn-primary flex-1" @click="onConfirm">确认完成</button>
			</template>
			<template v-if="role==='employer' && o.status===5 && !reviewed">
				<button class="btn-primary flex-1" @click="onReview">评价学生</button>
			</template>
		</view>
	</view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { orderApi } from '@/api/order'
	import { ORDER_STATUS_TEXT, ORDER_STATUS } from '@/utils/constants'
	import { formatMoney, formatDateTime } from '@/utils/format'
	import { confirm, toastSuccess, toastError } from '@/utils/ui'

	export default {
		data() { return { o: null, reviewed: false } },
		computed: {
			user() { return useUserStore() },
			role() { return this.user.activeRole },
			statusText() { return ORDER_STATUS_TEXT[this.o.status] || '—' },
			statusSub() {
				const map = {
					[ORDER_STATUS.WAIT_PAY]: '请尽快支付,超时订单会自动取消',
					[ORDER_STATUS.PAID]: '已支付,等待学生上岗打卡',
					[ORDER_STATUS.IN_PROGRESS]: '学生已开始工作',
					[ORDER_STATUS.WAIT_CONFIRM]: '学生已提交完成,等待您确认',
					[ORDER_STATUS.SETTLED]: '订单已完成,欢迎评价',
				}
				return map[this.o.status] || ''
			},
		},
		onLoad(q) { this.id = q.id; this.load() },
		methods: {
			formatMoney, formatTime: formatDateTime,
			preview(i) {
				uni.previewImage({ urls: this.o.work_proof, current: this.o.work_proof[i] })
			},
			async load() {
				try {
					const fn = this.role === 'student' ? orderApi.detail : orderApi.employerDetail
					this.o = await fn(this.id)
				} catch (e) { toastError('加载失败') }
			},
			async onPay() {
				try {
					this.o = await orderApi.pay(this.id, { method: 'mock_wechat' })
					toastSuccess('支付成功')
				} catch (e) {}
			},
			async onCheckin() {
				// Demo: use a placeholder URL. In production this is uni.chooseImage → upload.
				try {
					this.o = await orderApi.checkin(this.id, {
						work_proof: ['https://picsum.photos/seed/' + this.id + '/400/300'],
					})
					toastSuccess('已打卡')
				} catch (e) {}
			},
			async onComplete() {
				try {
					this.o = await orderApi.complete(this.id)
					toastSuccess('已提交,等待雇主确认')
				} catch (e) {}
			},
			async onConfirm() {
				const ok = await confirm('确认完成并结算?')
				if (!ok) return
				try {
					this.o = await orderApi.confirm(this.id)
					toastSuccess('已结算')
				} catch (e) {}
			},
			async onCancel() {
				uni.showModal({
					title: '取消订单',
					editable: true,
					placeholderText: '请输入取消原因',
					success: async r => {
						if (!r.confirm || !r.content) return
						try {
							const fn = this.role === 'student' ? orderApi.cancel : orderApi.employerCancel
							this.o = await fn(this.id, { reason: r.content })
							toastSuccess('已取消')
						} catch (e) {}
					},
				})
			},
			async onReview() {
				uni.showModal({
					title: '评价',
					editable: true,
					placeholderText: '给个评价(1-5星 + 文字)',
					success: async r => {
						if (!r.confirm) return
						const content = r.content || ''
						// 简单解析:取第一个数字字符作为评分
						const m = content.match(/[1-5]/)
						const rating = m ? Number(m[0]) : 5
						try {
							const fn = this.role === 'student' ? orderApi.review : orderApi.employerReview
							await fn(this.id, { rating, content: content.replace(/[1-5]/, '').trim() || '好评' })
							toastSuccess('已评价')
							this.reviewed = true
						} catch (e) {}
					},
				})
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.page { padding-bottom: 180rpx; }
	.status-bar {
		padding: $spacing-md; text-align: center; color: #fff;
		&.st-1 { background: $color-warning; }
		&.st-2 { background: $color-info; }
		&.st-3 { background: $brand-primary; }
		&.st-4 { background: $color-warning; }
		&.st-5 { background: $color-success; }
		&.st-6, &.st-7 { background: $text-placeholder; }
		.st-text { display: block; font-size: 40rpx; font-weight: 600; }
		.st-sub { display: block; font-size: $font-sm; opacity: 0.9; margin-top: 4rpx; }
	}
	.hero { margin: $spacing-sm; }
	.title { font-size: 36rpx; font-weight: 600; color: $text-primary; display: block; margin-bottom: $spacing-sm; }
	.amount-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: $spacing-sm; }
	.amount { font-size: 44rpx; color: $brand-primary; font-weight: 600; }
	.muted { color: $text-secondary; font-size: $font-sm; }
	.row { display: flex; padding: 12rpx 0; border-bottom: 2rpx solid $border-color-light; }
	.row:last-child { border-bottom: none; }
	.lbl { width: 160rpx; color: $text-secondary; font-size: $font-base; }
	.val { flex: 1; color: $text-primary; font-size: $font-base; }
	.section-title { font-size: $font-md; font-weight: 600; color: $text-primary; margin-bottom: $spacing-sm; }
	.proof-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: $spacing-xs; }
	.proof-img { width: 100%; height: 200rpx; border-radius: $radius-sm; }
	.footer {
		position: fixed; left: 0; right: 0; bottom: 0; padding: $spacing-sm $spacing-md;
		background: $bg-card; display: flex; gap: $spacing-sm; align-items: center;
		box-shadow: 0 -2rpx 16rpx rgba(0,0,0,0.06);
	}
	.flex-1 { flex: 1; }
</style>
