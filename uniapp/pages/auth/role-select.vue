<template>
	<view class="role-page">
		<view class="hero">
			<view class="logo">🎓</view>
			<view class="title">校园灵活用工</view>
			<view class="subtitle">选择身份开始</view>
		</view>

		<view class="cards">
			<view class="role-card" @click="pick('student')">
				<view class="emoji">🧑‍🎓</view>
				<view class="t">我是学生</view>
				<view class="d">找靠谱兼职 · 赚零花钱</view>
			</view>
			<view class="role-card emp" @click="pick('employer')">
				<view class="emoji">🏪</view>
				<view class="t">我是雇主</view>
				<view class="d">招靠谱学生 · 发布岗位</view>
			</view>
		</view>

		<view class="bottom">
			<text class="muted" @click="goLogin">已有账号?立即登录</text>
		</view>
	</view>
</template>

<script>
	import { setActiveRole } from '@/utils/auth'

	export default {
		methods: {
			pick(role) {
				setActiveRole(role)
				uni.redirectTo({ url: '/pages/auth/register?role=' + role })
			},
			goLogin() {
				uni.redirectTo({ url: '/pages/auth/login' })
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.role-page { min-height: 100vh; background: $bg-page; padding: 100rpx $spacing-md $spacing-xl; }
	.hero { text-align: center; margin-bottom: 80rpx; }
	.logo { font-size: 140rpx; line-height: 1; }
	.title { font-size: 48rpx; font-weight: 600; color: $text-primary; margin-top: 24rpx; }
	.subtitle { font-size: $font-base; color: $text-secondary; margin-top: 12rpx; }
	.cards { display: flex; flex-direction: column; gap: $spacing-md; }
	.role-card {
		background: $bg-card; border-radius: $radius-lg; padding: $spacing-lg;
		box-shadow: $shadow-md; text-align: center; position: relative; overflow: hidden;
		&::before {
			content: ''; position: absolute; top: 0; left: 0; right: 0; height: 8rpx;
			background: linear-gradient(90deg, $brand-primary 0%, $brand-accent 100%);
		}
		&.emp::before { background: linear-gradient(90deg, $brand-secondary 0%, $brand-accent 100%); }
		.emoji { font-size: 100rpx; line-height: 1; }
		.t { font-size: $font-xl; font-weight: 600; color: $text-primary; margin-top: $spacing-sm; }
		.d { font-size: $font-sm; color: $text-secondary; margin-top: $spacing-xs; }
	}
	.bottom { text-align: center; margin-top: $spacing-xl; }
	.muted { color: $brand-primary; font-size: $font-base; }
</style>
