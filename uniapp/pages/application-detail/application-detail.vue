<template>
	<view class="page app-detail" v-if="a">
		<view class="card hero">
			<view class="row"><text class="lbl">岗位</text><text class="val">{{ a.job_title }}</text></view>
			<view class="row"><text class="lbl">状态</text><text class="tag" :class="statusTag(a.status)">{{ statusText(a.status) }}</text></view>
			<view class="row" v-if="role==='student'"><text class="lbl">报名人</text><text class="val">{{ user.user.nickname || user.user.username }}</text></view>
			<view class="row" v-else><text class="lbl">报名人</text><text class="val">{{ a.student_name }} · {{ a.student_school }}</text></view>
			<view class="row"><text class="lbl">联系电话</text><text class="val">{{ a.contact_phone || '—' }}</text></view>
			<view class="row"><text class="lbl">留言</text><text class="val">{{ a.message || '—' }}</text></view>
			<view class="row" v-if="a.audit_remark"><text class="lbl">审核意见</text><text class="val">{{ a.audit_remark }}</text></view>
			<view class="row"><text class="lbl">报名时间</text><text class="val">{{ formatTime(a.created_at) }}</text></view>
		</view>

		<!-- 学生操作 -->
		<view v-if="role==='student' && a.status===1" class="footer">
			<button class="btn-ghost" @click="onCancel">取消报名</button>
		</view>

		<!-- 雇主操作 -->
		<view v-if="role==='employer' && a.status===1" class="footer">
			<button class="btn-secondary" @click="onAudit(3)">拒绝</button>
			<button class="btn-primary flex-1" @click="onAudit(2)">通过并录用</button>
		</view>
		<view v-if="role==='employer' && a.status===2" class="footer">
			<button class="btn-primary flex-1" @click="onHire">录用并下单</button>
		</view>
	</view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { jobApi } from '@/api/job'
	import { APP_STATUS_TEXT, APP_STATUS_TAG } from '@/utils/constants'
	import { formatDateTime } from '@/utils/format'
	import { confirm, toastSuccess, toastError } from '@/utils/ui'

	export default {
		data() { return { a: null } },
		computed: {
			user() { return useUserStore() },
			role() { return this.user.activeRole },
		},
		onLoad(q) { this.id = q.id; this.load() },
		methods: {
			formatTime: formatDateTime,
			statusText: (s) => APP_STATUS_TEXT[s] || '—',
			statusTag: (s) => APP_STATUS_TAG[s] || 'tag-info',
			async load() {
				try {
					// Student: detail API. Employer: comes from list, simpler to
					// fetch the underlying job's applications and find ours.
					if (this.role === 'student') {
						this.a = await jobApi.applicationDetail(this.id)
					} else {
						// Find by scanning — small dataset in the template.
						const my = await jobApi.myJobs({ page: 1, size: 100 })
						const jobs = (my && my.list) || []
						for (const j of jobs) {
							const r = await jobApi.jobApplications(j.id, { page: 1, size: 100 })
							const found = ((r && r.list) || []).find(x => x.id === Number(this.id))
							if (found) { this.a = found; break }
						}
					}
				} catch (e) { toastError('加载失败') }
			},
			async onCancel() {
				const ok = await confirm('确认取消报名?')
				if (!ok) return
				try {
					await jobApi.cancelApplication(this.id)
					toastSuccess('已取消')
					setTimeout(() => uni.navigateBack(), 600)
				} catch (e) {}
			},
			async onAudit(action) {
				const ok = await confirm(action === 2 ? '通过并录用该学生?' : '拒绝该报名?')
				if (!ok) return
				try {
					await jobApi.auditApplication(this.id, { action, remark: action === 2 ? '欢迎加入' : '暂不匹配' })
					toastSuccess(action === 2 ? '已通过' : '已拒绝')
					this.load()
				} catch (e) {}
			},
			async onHire() {
				// Ask for amount first.
				uni.showModal({
					title: '设置订单金额',
					editable: true,
					placeholderText: '输入金额(元)',
					success: async r => {
						if (!r.confirm) return
						const amount = parseFloat(r.content)
						if (!amount || amount <= 0) return toastError('金额无效')
						try {
							const o = await jobApi.hire(this.id, { amount })
							toastSuccess('已生成订单')
							setTimeout(() => {
								uni.redirectTo({ url: '/pages/order-detail/order-detail?id=' + o.id })
							}, 600)
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
	.hero { margin: $spacing-sm; }
	.row { display: flex; padding: 16rpx 0; border-bottom: 2rpx solid $border-color-light; }
	.row:last-child { border-bottom: none; }
	.lbl { width: 160rpx; color: $text-secondary; font-size: $font-base; }
	.val { flex: 1; color: $text-primary; font-size: $font-base; }
	.footer {
		position: fixed; left: 0; right: 0; bottom: 0; padding: $spacing-sm $spacing-md;
		background: $bg-card; display: flex; gap: $spacing-sm; align-items: center;
		box-shadow: 0 -2rpx 16rpx rgba(0,0,0,0.06);
	}
	.flex-1 { flex: 1; }
</style>
