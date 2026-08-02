<template>
	<view class="page register-page">
		<view class="hero">
			<view class="emoji">{{ roleEmoji }}</view>
			<view class="title">{{ roleTitle }}注册</view>
			<view class="subtitle">注册即登录,仅需用户名+密码</view>
		</view>

		<view class="form card">
			<view class="field">
				<text class="label">用户名</text>
				<input class="input" v-model="form.username" placeholder="3-64位,字母数字下划线" maxlength="64" />
			</view>
			<view class="field">
				<text class="label">密码</text>
				<input class="input" v-model="form.password" type="password" placeholder="6-128位" maxlength="128" />
			</view>
			<view class="field">
				<text class="label">昵称(可选)</text>
				<input class="input" v-model="form.nickname" maxlength="64" />
			</view>
			<view class="field">
				<text class="label">手机号(可选)</text>
				<input class="input" v-model="form.phone" maxlength="32" type="number" />
			</view>

			<button class="btn-primary submit" :disabled="loading" @click="onSubmit">
				{{ loading ? '注册中...' : '注册并登录' }}
			</button>
			<view class="hint">
				<text class="muted" @click="goLogin">已有账号?返回登录</text>
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
				role: 'student',
				form: { username: '', password: '', nickname: '', phone: '' },
				loading: false,
			}
		},
		computed: {
			roleEmoji() { return this.role === 'employer' ? '🏪' : '🧑‍🎓' },
			roleTitle() { return this.role === 'employer' ? '雇主' : '学生' },
		},
		onLoad(query) {
			if (query && query.role) this.role = query.role
		},
		methods: {
			async onSubmit() {
				if (!this.form.username || !this.form.password) {
					return toastError('请填写用户名和密码')
				}
				this.loading = true
				try {
					const u = useUserStore()
					await u.register({ ...this.form, user_type: this.role })
					toastSuccess('注册成功')
					setTimeout(() => uni.switchTab({ url: '/pages/index/index' }), 500)
				} catch (e) {} finally {
					this.loading = false
				}
			},
			goLogin() { uni.redirectTo({ url: '/pages/auth/login' }) },
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.register-page { padding-top: 40rpx; }
	.hero { text-align: center; margin-bottom: $spacing-md; }
	.emoji { font-size: 96rpx; }
	.title { font-size: 36rpx; font-weight: 600; color: $text-primary; margin-top: 12rpx; }
	.subtitle { font-size: $font-sm; color: $text-secondary; margin-top: 4rpx; }
	.form { padding: $spacing-md; }
	.field { padding: 14rpx 0; border-bottom: 2rpx solid $border-color-light; }
	.field:last-of-type { border-bottom: none; }
	.label { font-size: $font-sm; color: $text-secondary; }
	.input { display: block; margin-top: 6rpx; font-size: $font-lg; color: $text-primary; }
	.submit { width: 100%; margin-top: $spacing-md; }
	.hint { text-align: center; margin-top: $spacing-md; }
	.muted { color: $brand-primary; font-size: $font-sm; }
</style>
