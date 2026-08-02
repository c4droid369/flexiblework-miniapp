<template>
	<view class="page login-page">
		<view class="hero">
			<view class="emoji">👋</view>
			<view class="title">欢迎回来</view>
			<view class="subtitle">登录后开启你的校园兼职之旅</view>
		</view>

		<view class="form card">
			<view class="field">
				<text class="label">用户名</text>
				<input class="input" v-model="form.username" placeholder="请输入用户名" maxlength="64" />
			</view>
			<view class="field">
				<text class="label">密码</text>
				<input class="input" v-model="form.password" type="password" placeholder="请输入密码" maxlength="128" />
			</view>
			<button class="btn-primary submit" :disabled="loading" @click="onSubmit">
				{{ loading ? '登录中...' : '登 录' }}
			</button>
			<view class="hint">
				<text class="muted" @click="goRegister">还没有账号?立即注册</text>
			</view>
		</view>

		<view class="demo">
			<view class="title">体验账号</view>
			<view class="acc" @click="fill('demo_student', 'demo123')">
				<text class="role">学生</text>
				<text class="name">demo_student / demo123</text>
			</view>
			<view class="acc" @click="fill('demo_employer', 'demo123')">
				<text class="role">雇主</text>
				<text class="name">demo_employer / demo123</text>
			</view>
			<view class="acc" @click="fill('admin', 'admin123')">
				<text class="role">管理</text>
				<text class="name">admin / admin123</text>
			</view>
		</view>
	</view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { toastError, toastSuccess } from '@/utils/ui'

	export default {
		data() {
			return {
				form: { username: '', password: '' },
				loading: false,
			}
		},
		methods: {
			fill(u, p) {
				this.form.username = u
				this.form.password = p
			},
			async onSubmit() {
				if (!this.form.username || !this.form.password) {
					return toastError('请填写完整')
				}
				this.loading = true
				try {
					const u = useUserStore()
					await u.login(this.form)
					toastSuccess('登录成功')
					setTimeout(() => {
						if (u.userType === 'admin') {
							// Admins don't use the mini-program tabbar in this
							// template; they use the Vue admin panel. Bounce
							// to a friendly "this is the mini-program" page.
							uni.showModal({
								title: '管理员账号',
								content: '管理员请前往 admin/ 前端管理后台,本小程序仅服务学生/雇主。',
								showCancel: false,
								success: () => u.logout().then(() => uni.reLaunch({ url: '/pages/auth/role-select' })),
							})
						} else {
							uni.switchTab({ url: '/pages/index/index' })
						}
					}, 500)
				} catch (e) {
					// toast handled by request layer
				} finally {
					this.loading = false
				}
			},
			goRegister() { uni.redirectTo({ url: '/pages/auth/role-select' }) },
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.login-page { padding-top: 60rpx; }
	.hero { text-align: center; margin-bottom: $spacing-lg; }
	.emoji { font-size: 100rpx; }
	.title { font-size: 40rpx; font-weight: 600; color: $text-primary; margin-top: 16rpx; }
	.subtitle { font-size: $font-sm; color: $text-secondary; margin-top: 8rpx; }
	.form { padding: $spacing-md; }
	.field { padding: 16rpx 0; border-bottom: 2rpx solid $border-color-light; }
	.field:last-of-type { border-bottom: none; }
	.label { font-size: $font-sm; color: $text-secondary; }
	.input { display: block; margin-top: 8rpx; font-size: $font-lg; color: $text-primary; }
	.submit { width: 100%; margin-top: $spacing-md; }
	.hint { text-align: center; margin-top: $spacing-md; }
	.muted { color: $brand-primary; font-size: $font-sm; }
	.demo {
		margin-top: $spacing-lg; background: $brand-primary-bg; border-radius: $radius-md;
		padding: $spacing-md;
		.title { font-size: $font-sm; color: $text-secondary; margin-bottom: $spacing-sm; }
		.acc {
			display: flex; align-items: center; gap: $spacing-sm;
			background: $bg-card; border-radius: $radius-sm; padding: 12rpx $spacing-sm;
			margin-bottom: $spacing-xs;
			.role { background: $brand-primary; color: #fff; font-size: $font-xs; padding: 2rpx 12rpx; border-radius: $radius-pill; }
			.name { font-size: $font-sm; color: $text-regular; }
		}
	}
</style>
