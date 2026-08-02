<template>
	<view class="page cert-page">
		<view class="tabs">
			<view class="tab" :class="{ active: tab==='profile' }" @click="tab='profile'">个人资料</view>
			<view class="tab" :class="{ active: tab==='cert' }" @click="tab='cert'">实名认证</view>
		</view>

		<!-- Profile form (both roles) -->
		<view v-if="tab==='profile'" class="form card">
			<template v-if="role==='student'">
				<view class="field"><text class="label">真实姓名</text><input class="input" v-model="form.real_name" maxlength="64" /></view>
				<view class="field"><text class="label">性别</text>
					<view class="radios">
						<view class="radio" :class="{ active: form.gender===0 }" @click="form.gender=0">未填</view>
						<view class="radio" :class="{ active: form.gender===1 }" @click="form.gender=1">男</view>
						<view class="radio" :class="{ active: form.gender===2 }" @click="form.gender=2">女</view>
					</view>
				</view>
				<view class="field"><text class="label">学校</text><input class="input" v-model="form.school" maxlength="128" placeholder="示例大学" /></view>
				<view class="field"><text class="label">学院</text><input class="input" v-model="form.college" maxlength="128" /></view>
				<view class="field"><text class="label">专业</text><input class="input" v-model="form.major" maxlength="128" /></view>
				<view class="field"><text class="label">学号</text><input class="input" v-model="form.student_no" maxlength="64" /></view>
				<view class="field"><text class="label">个人简介</text><textarea class="textarea" v-model="form.bio" maxlength="500" placeholder="一句话介绍自己"></textarea></view>
			</template>
			<template v-else>
				<view class="field"><text class="label">公司/店铺名称</text><input class="input" v-model="form.company_name" maxlength="128" /></view>
				<view class="field"><text class="label">联系人</text><input class="input" v-model="form.contact_name" maxlength="64" /></view>
				<view class="field"><text class="label">联系电话</text><input class="input" v-model="form.contact_phone" maxlength="32" /></view>
				<view class="field"><text class="label">邮箱</text><input class="input" v-model="form.contact_email" maxlength="128" /></view>
				<view class="field"><text class="label">行业</text><input class="input" v-model="form.industry" maxlength="64" /></view>
				<view class="field"><text class="label">规模</text><input class="input" v-model="form.company_size" maxlength="32" placeholder="如 1-10" /></view>
				<view class="field"><text class="label">地址</text><input class="input" v-model="form.company_address" maxlength="255" /></view>
				<view class="field"><text class="label">简介</text><textarea class="textarea" v-model="form.intro" maxlength="500"></textarea></view>
			</template>
			<button class="btn-primary save" :disabled="saving" @click="onSave">{{ saving ? '保存中...' : '保存资料' }}</button>
		</view>

		<!-- Cert form (both roles) -->
		<view v-else class="form card">
			<view v-if="certStatus===2" class="state-card approved">
				<text class="big">✅</text>
				<text class="t">已通过认证</text>
				<text class="muted">你已具备完整的{{ role==='student' ? '学生' : '雇主' }}权限</text>
			</view>
			<view v-else-if="certStatus===1" class="state-card pending">
				<text class="big">⏳</text>
				<text class="t">审核中</text>
				<text class="muted">管理端会在 1-2 个工作日内审核,请耐心等待</text>
			</view>
			<view v-else>
				<text class="muted intro">请{{ role==='student' ? '上传身份证正反面 + 学生证' : '填写以下信息并上传营业执照' }},提交后进入审核。</text>
				<view v-if="role==='student'">
					<image-uploader v-model="form.id_card_front" label="身份证正面" />
					<image-uploader v-model="form.id_card_back" label="身份证反面" />
					<image-uploader v-model="form.student_card" label="学生证" />
				</view>
				<view v-else>
					<view class="field"><text class="label">公司名称</text><input class="input" v-model="certForm.company_name" maxlength="128" /></view>
					<view class="field"><text class="label">营业执照号</text><input class="input" v-model="certForm.business_license_no" maxlength="64" /></view>
					<view class="field"><text class="label">联系人</text><input class="input" v-model="certForm.contact_name" maxlength="64" /></view>
					<view class="field"><text class="label">联系电话</text><input class="input" v-model="certForm.contact_phone" maxlength="32" /></view>
					<image-uploader v-model="certForm.business_license_img" label="营业执照" />
				</view>
				<button class="btn-primary save" :disabled="saving" @click="onSubmitCert">{{ saving ? '提交中...' : '提交认证' }}</button>
			</view>
		</view>
	</view>
</template>

<script>
	import { useUserStore } from '@/store/user'
	import { profileApi } from '@/api/profile'
	import { uploadApi } from '@/api/upload'
	import { CERT_STATUS_TEXT } from '@/utils/constants'
	import { toastSuccess, toastError } from '@/utils/ui'

	// Inline image uploader — minimal component, no separate file.
	const ImageUploader = {
		props: ['modelValue', 'label'],
		emits: ['update:modelValue'],
		methods: {
			async pick() {
				try {
					const r = await uni.chooseImage({ count: 1, sizeType: ['compressed'] })
					const path = r.tempFilePaths[0]
					uni.showLoading({ title: '上传中...' })
					const data = await uploadApi.image(path)
					this.$emit('update:modelValue', data.url)
				} catch (e) { toastError('上传失败') }
				finally { uni.hideLoading() }
			},
		},
		template: `
			<view class="uploader" @click="pick">
				<image v-if="modelValue" class="img" :src="modelValue" mode="aspectFill" />
				<view v-else class="placeholder">
					<text class="plus">+</text>
					<text class="lbl">{{ label || '上传图片' }}</text>
				</view>
			</view>
		`,
	}

	export default {
		components: { ImageUploader },
		data() {
			return {
				tab: 'profile',
				form: {},
				certForm: { business_license_no: '', company_name: '', contact_name: '', contact_phone: '' },
				certStatus: 0,
				saving: false,
			}
		},
		computed: {
			user() { return useUserStore() },
			role() { return this.user.activeRole },
		},
		onLoad(q) {
			if (q && q.tab) this.tab = q.tab
			this.load()
		},
		methods: {
			async load() {
				try {
					if (this.role === 'student') {
						const p = await profileApi.getStudent()
						this.form = { gender: 0, ...(p || {}) }
						this.certStatus = (p && p.cert_status) || 0
					} else {
						const p = await profileApi.getEmployer()
						this.form = { ...(p || {}) }
						this.certStatus = (p && p.cert_status) || 0
					}
				} catch (e) {}
			},
			async onSave() {
				this.saving = true
				try {
					const fn = this.role === 'student' ? profileApi.updateStudent : profileApi.updateEmployer
					const cleaned = this.clean(this.form)
					await fn(cleaned)
					toastSuccess('保存成功')
					this.load()
				} catch (e) {} finally { this.saving = false }
			},
			async onSubmitCert() {
				this.saving = true
				try {
					if (this.role === 'student') {
						if (!this.form.id_card_front || !this.form.id_card_back || !this.form.student_card) {
							return toastError('请上传所有图片')
						}
						await profileApi.submitStudentCert({
							id_card_front: this.form.id_card_front,
							id_card_back: this.form.id_card_back,
							student_card: this.form.student_card,
						})
					} else {
						if (!this.certForm.company_name || !this.certForm.business_license_no ||
							!this.certForm.contact_name || !this.certForm.contact_phone ||
							!this.certForm.business_license_img) {
							return toastError('请填写完整并上传营业执照')
						}
						await profileApi.submitEmployerCert(this.certForm)
					}
					toastSuccess('已提交,等待审核')
					this.load()
				} catch (e) {} finally { this.saving = false }
			},
			clean(o) {
				// Backend uses pointer fields; we send all present keys. Empty
				// string fields are dropped to avoid overwriting with blanks.
				const out = {}
				Object.keys(o).forEach(k => {
					const v = o[k]
					if (v !== '' && v !== null && v !== undefined) out[k] = v
				})
				return out
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.tabs { display: flex; padding: $spacing-sm $spacing-md; gap: $spacing-sm; background: $bg-card; }
	.tab {
		flex: 1; text-align: center; padding: 20rpx 0; font-size: $font-base; color: $text-secondary;
		border-bottom: 4rpx solid transparent;
		&.active { color: $brand-primary; border-bottom-color: $brand-primary; font-weight: 600; }
	}
	.form { margin: $spacing-sm; }
	.field { padding: 20rpx 0; border-bottom: 2rpx solid $border-color-light; }
	.field:last-of-type { border-bottom: none; }
	.label { font-size: $font-sm; color: $text-secondary; }
	.input { display: block; margin-top: 8rpx; font-size: $font-lg; color: $text-primary; }
	.textarea { display: block; margin-top: 8rpx; font-size: $font-base; color: $text-primary; width: 100%; height: 120rpx; }
	.radios { display: flex; gap: $spacing-sm; margin-top: 8rpx; }
	.radio {
		padding: 12rpx 32rpx; background: $bg-page; border-radius: $radius-pill; font-size: $font-base; color: $text-regular;
		&.active { background: $brand-primary-bg; color: $brand-primary; }
	}
	.save { width: 100%; margin-top: $spacing-md; }
	.muted { color: $text-secondary; font-size: $font-sm; }
	.form-row { display: flex; flex-direction: column; gap: $spacing-md; }
	.approved, .pending { text-align: center; padding: $spacing-xl 0; }
	.approved .big { font-size: 96rpx; }
	.pending .big { font-size: 96rpx; }
	.t { display: block; font-size: $font-lg; font-weight: 600; margin-top: $spacing-sm; }
	.card-inner { background: $bg-card; }
</style>
