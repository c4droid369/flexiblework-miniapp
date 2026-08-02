<script>
	import { useUserStore } from '@/store/user'
	import { useAppStore } from '@/store/app'

	export default {
		onLaunch: function() {
			console.log('[campus-gig] app launch')
			// Restore session from local storage on cold start.
			const user = useUserStore()
			user.bootstrap()
			// App-wide config (categories etc.).
			const app = useAppStore()
			app.bootstrap()
		},
		onShow: function() {
			console.log('[campus-gig] app show')
		},
		onHide: function() {
			console.log('[campus-gig] app hide')
		}
	}
</script>

<style lang="scss">
	/* Global resets + brand styling. */
	@import 'uni.scss';

	page {
		background-color: $bg-page;
		color: $text-primary;
		font-size: $font-base;
		line-height: $line-normal;
		font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
	}

	view, text {
		box-sizing: border-box;
	}

	/* Reusable utility classes */
	.flex { display: flex; }
	.flex-col { display: flex; flex-direction: column; }
	.flex-center { display: flex; align-items: center; justify-content: center; }
	.flex-between { display: flex; align-items: center; justify-content: space-between; }
	.flex-1 { flex: 1; }

	.text-primary { color: $text-primary; }
	.text-regular { color: $text-regular; }
	.text-secondary { color: $text-secondary; }
	.text-placeholder { color: $text-placeholder; }
	.text-brand { color: $brand-primary; }
	.text-success { color: $color-success; }
	.text-warning { color: $color-warning; }
	.text-danger { color: $color-danger; }

	.text-xs { font-size: $font-xs; }
	.text-sm { font-size: $font-sm; }
	.text-base { font-size: $font-base; }
	.text-lg { font-size: $font-lg; }
	.text-xl { font-size: $font-xl; }

	.font-bold { font-weight: 600; }
	.font-medium { font-weight: 500; }

	.bg-card { background-color: $bg-card; }
	.bg-page { background-color: $bg-page; }

	/* Primary button — used everywhere. */
	.btn-primary {
		background: linear-gradient(135deg, $brand-primary 0%, $brand-primary-dark 100%);
		color: #fff;
		border-radius: $radius-pill;
		font-size: $font-md;
		font-weight: 500;
		padding: 20rpx 48rpx;
		border: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		&:active { opacity: 0.85; }
		&[disabled], &.is-disabled { opacity: 0.5; }
	}

	.btn-secondary {
		background: $bg-card;
		color: $brand-primary;
		border: 2rpx solid $brand-primary;
		border-radius: $radius-pill;
		font-size: $font-md;
		padding: 18rpx 46rpx;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		&:active { background: $brand-primary-bg; }
	}

	.btn-ghost {
		background: transparent;
		color: $text-regular;
		font-size: $font-base;
		padding: 16rpx 24rpx;
		border: none;
		&:active { opacity: 0.6; }
	}

	/* Card */
	.card {
		background: $bg-card;
		border-radius: $radius-md;
		padding: $spacing-md;
		box-shadow: $shadow-sm;
	}

	/* Tag / status pill */
	.tag {
		display: inline-flex;
		align-items: center;
		padding: 4rpx 16rpx;
		font-size: $font-xs;
		border-radius: $radius-pill;
		background: $bg-hover;
		color: $text-regular;
	}
	.tag-primary { background: $brand-primary-bg; color: $brand-primary; }
	.tag-success { background: $brand-secondary-bg; color: $color-success; }
	.tag-warning { background: #FFF7E6; color: $color-warning; }
	.tag-danger { background: #FFECE8; color: $color-danger; }
	.tag-info { background: #F0F4FA; color: $color-info; }

	/* Section title with optional right slot */
	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: $spacing-md $spacing-sm $spacing-sm;
		.title { font-size: $font-lg; font-weight: 600; color: $text-primary; }
		.more { font-size: $font-sm; color: $text-secondary; }
	}

	/* Empty state */
	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: $spacing-xl 0;
		color: $text-placeholder;
		font-size: $font-sm;
		.emoji { font-size: 80rpx; margin-bottom: $spacing-sm; }
	}

	/* Page container (consistent horizontal padding) */
	.page {
		min-height: 100vh;
		padding: $spacing-sm $spacing-md;
	}

	/* Safe area at bottom for tabbar pages */
	.safe-bottom { padding-bottom: env(safe-area-inset-bottom); }
</style>
