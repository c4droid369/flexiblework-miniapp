<template>
	<view class="page settings-page">
		<view class="card">
			<view class="section-title">API 服务器地址</view>
			<text class="muted hint">
				真机调试时把 localhost 换成电脑局域网 IP(在电脑上 `ipconfig` 查,形如 192.168.1.20)。\n
				Android 模拟器用 10.0.2.2;iOS 模拟器 / 微信开发者工具用 localhost。
			</text>
			<view class="field">
				<text class="label">当前地址</text>
				<text class="current">{{ currentUrl }}</text>
			</view>
			<view class="field">
				<text class="label">新地址</text>
				<input class="input" v-model="draft" :placeholder="defaultUrl" />
			</view>
			<view class="actions">
				<button class="btn-primary" :disabled="saving" @click="onSave">{{ saving ? '保存中...' : '保存并测试' }}</button>
				<button class="btn-ghost" @click="onReset">恢复默认</button>
			</view>
			<view v-if="testResult" class="result" :class="testOk ? 'ok' : 'err'">
				<text>{{ testResult }}</text>
			</view>
		</view>

		<view class="card">
			<view class="section-title">预设场景</view>
			<view class="presets">
				<view class="preset" @click="draft='http://localhost:8080'">微信开发者工具<br/><text class="muted">localhost</text></view>
				<view class="preset" @click="draft='http://10.0.2.2:8080'">Android 模拟器<br/><text class="muted">10.0.2.2</text></view>
				<view class="preset" @click="draft='http://127.0.0.1:8080'">本机回环<br/><text class="muted">127.0.0.1</text></view>
			</view>
		</view>

		<view class="card">
			<view class="section-title">说明</view>
			<text class="muted">
				后端如果在 docker compose 跑,默认绑定 0.0.0.0:8080,所以同网段设备都能访问。
				改完地址后,所有 API 请求(登录、上传、订单等)都会立即用新地址,不需要重启。
			</text>
		</view>
	</view>
</template>

<script setup>
	import { ref, computed } from 'vue'
	import { getApiBaseMeta, setApiBase, getFullApiBase } from '@/utils/api-base'
	import { toastSuccess, toastError } from '@/utils/ui'

	const draft = ref('')
	const saving = ref(false)
	const testResult = ref('')
	const testOk = ref(false)
	const meta = ref(getApiBaseMeta())

	const currentUrl = computed(() => meta.value.url)
	const defaultUrl = computed(() => meta.value.default)

	async function onSave() {
		const url = draft.value.trim()
		if (!url) return toastError('请输入新地址')
		if (!/^https?:\/\/.+/.test(url)) return toastError('地址必须以 http:// 或 https:// 开头')
		setApiBase(url)
		meta.value = getApiBaseMeta()
		testResult.value = ''
		await runTest()
	}

	function onReset() {
		setApiBase('')
		meta.value = getApiBaseMeta()
		draft.value = ''
		testResult.value = ''
		toastSuccess('已恢复默认')
	}

	async function runTest() {
		saving.value = true
		try {
			const r = await new Promise((resolve, reject) => {
				uni.request({
					url: getFullApiBase() + '/categories',
					method: 'GET',
					timeout: 8000,
					success: resolve,
					fail: reject,
				})
			})
			const body = typeof r.data === 'string' ? JSON.parse(r.data || '{}') : r.data
			if (r.statusCode === 200 && body && body.code === 0) {
				const n = Array.isArray(body.data) ? body.data.length : '?'
				testResult.value = '✓ 连接成功,返回 ' + n + ' 个分类'
				testOk.value = true
				toastSuccess('保存成功,连接正常')
			} else {
				testResult.value = '✗ 后端返回错误: ' + (body && body.message) || ('HTTP ' + r.statusCode)
				testOk.value = false
			}
		} catch (e) {
			testOk.value = false
			testResult.value = '✗ 连接失败: ' + (e.errMsg || e.message || '网络错误')
		} finally {
			saving.value = false
		}
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.settings-page { padding-top: 0; }
	.card { margin: $spacing-sm; }
	.section-title { font-size: $font-md; font-weight: 600; color: $text-primary; margin-bottom: $spacing-sm; }
	.hint { display: block; line-height: $line-relaxed; white-space: pre-line; margin-bottom: $spacing-md; }
	.field { padding: 16rpx 0; border-bottom: 2rpx solid $border-color-light; }
	.field:last-of-type { border-bottom: none; }
	.label { font-size: $font-sm; color: $text-secondary; }
	.current { display: block; margin-top: 6rpx; font-size: $font-lg; color: $brand-primary; word-break: break-all; }
	.input { display: block; margin-top: 8rpx; font-size: $font-lg; color: $text-primary; padding: 16rpx 24rpx; background: $bg-page; border-radius: $radius-sm; }
	.actions { display: flex; gap: $spacing-sm; margin-top: $spacing-md; }
	.actions .btn-primary { flex: 1; }
	.btn-ghost { padding: 16rpx 32rpx; }
	.result { margin-top: $spacing-md; padding: 16rpx 24rpx; border-radius: $radius-sm; font-size: $font-sm; }
	.result.ok { background: $brand-secondary-bg; color: $color-success; }
	.result.err { background: #FFECE8; color: $color-danger; }
	.presets { display: grid; grid-template-columns: repeat(3, 1fr); gap: $spacing-sm; }
	.preset { padding: 20rpx 16rpx; background: $bg-page; border-radius: $radius-md; text-align: center; font-size: $font-sm; color: $text-primary; }
	.preset:active { background: $brand-primary-bg; }
	.muted { color: $text-secondary; }
	.muted.hint, .card > .muted { line-height: $line-relaxed; }
</style>
