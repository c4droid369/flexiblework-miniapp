<template>
	<view class="page jobs-page">
		<!-- 顶部欢迎条 -->
		<view class="header">
			<view class="greet">
				<text class="hi">Hi, {{ nickname }}</text>
				<text class="sub">想找点什么兼职?</text>
			</view>
			<view class="msg-btn" @click="goMessages">
				<text class="ico">🔔</text>
			</view>
		</view>

		<!-- 搜索条 -->
		<view class="search" @click="onSearch">
			<text class="ico">🔍</text>
			<text class="placeholder">搜索岗位 / 地点 / 关键字</text>
		</view>

		<!-- 分类横滚条 -->
		<scroll-view scroll-x class="cats" :show-scrollbar="false">
			<view class="cat" :class="{ active: !currentCat }" @click="pickCat(null)">
				<text>全部</text>
			</view>
			<view v-for="c in categories" :key="c.id" class="cat" :class="{ active: currentCat === c.id }" @click="pickCat(c.id)">
				<text>{{ c.name }}</text>
			</view>
		</scroll-view>

		<!-- 列表 -->
		<view class="list">
			<job-card v-for="j in list" :key="j.id" :job="j" />
			<empty-state v-if="!loading && list.length === 0" emoji="🍃" text="没有匹配的岗位" />
			<view v-if="loading" class="loading"><text>加载中...</text></view>
		</view>
	</view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { useAppStore } from '@/store/app'
	import { jobApi } from '@/api/job'
	import JobCard from '@/components/job-card/job-card.vue'
	import EmptyState from '@/components/empty-state/empty-state.vue'

	export default {
		components: { JobCard, EmptyState },
		data() {
			return {
				keyword: '',
				currentCat: null,
				list: [],
				loading: false,
				page: 1,
				hasMore: true,
			}
		},
		computed: {
			nickname() {
				const u = useUserStore()
				return u.user && (u.user.nickname || u.user.username) || '同学'
			},
			categories() {
				const app = useAppStore()
				return app.activeCategories
			},
		},
		onLoad() {
			const app = useAppStore()
			app.loadCategories()
		},
		onShow() {
			// If we were routed away (e.g. to detail) and came back, refresh.
			this.refresh()
		},
		onPullDownRefresh() {
			this.refresh().finally(() => uni.stopPullDownRefresh())
		},
		onReachBottom() {
			if (this.hasMore && !this.loading) this.loadMore()
		},
		methods: {
			goMessages() { uni.navigateTo({ url: '/pages/messages/messages' }) },
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
			pickCat(id) {
				this.currentCat = id
				this.refresh()
			},
			async refresh() {
				this.page = 1
				this.hasMore = true
				await this.fetch()
			},
			async loadMore() {
				this.page++
				await this.fetch(true)
			},
			async fetch(append = false) {
				this.loading = true
				try {
					const params = { page: this.page, size: 10 }
					if (this.currentCat) params.category_id = this.currentCat
					if (this.keyword) params.keyword = this.keyword
					const res = await jobApi.list(params)
					const arr = (res && res.list) || []
					if (append) this.list = this.list.concat(arr)
					else this.list = arr
					this.hasMore = arr.length >= 10
				} catch (e) {} finally {
					this.loading = false
				}
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.jobs-page { padding-top: 0; }
	.header {
		display: flex; align-items: center; justify-content: space-between;
		background: linear-gradient(135deg, $brand-primary 0%, $brand-primary-dark 100%);
		color: #fff; padding: 40rpx $spacing-md $spacing-lg; border-bottom-left-radius: $radius-xl; border-bottom-right-radius: $radius-xl;
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
</style>
