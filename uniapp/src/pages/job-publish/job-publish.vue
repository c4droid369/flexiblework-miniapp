<template>
	<view class="page publish-page">
		<view class="form card">
			<view class="field">
				<text class="label">岗位名称 *</text>
				<input class="input" v-model="form.title" maxlength="128" placeholder="如:周末饮品店店员" />
			</view>
			<view class="field">
				<text class="label">分类 *</text>
				<view class="radios">
					<view v-for="c in categories" :key="c.id" class="radio" :class="{ active: form.category_id===c.id }" @click="form.category_id=c.id">
						<text>{{ c.name }}</text>
					</view>
				</view>
			</view>
			<view class="field">
				<text class="label">薪资方式 *</text>
				<view class="radios">
					<view v-for="(t, k) in salaryTypes" :key="k" class="radio" :class="{ active: form.salary_type===Number(k) }" @click="form.salary_type=Number(k)">
						<text>{{ t }}</text>
					</view>
				</view>
			</view>
			<view class="field">
				<text class="label">薪资范围(元) *</text>
				<view class="row">
					<input class="input small" v-model.number="form.salary_min" type="number" placeholder="最低" />
					<text class="dash">—</text>
					<input class="input small" v-model.number="form.salary_max" type="number" placeholder="最高" />
					<text class="unit">元</text>
				</view>
			</view>
			<view class="field">
				<text class="label">工作地点 *</text>
				<input class="input" v-model="form.location" maxlength="255" placeholder="如:学校东门商业街 12 号" />
			</view>
			<view class="field">
				<text class="label">招募人数 *</text>
				<input class="input" v-model.number="form.recruit_count" type="number" placeholder="如 2" />
			</view>
			<view class="field">
				<text class="label">性别要求</text>
				<view class="radios">
					<view class="radio" :class="{ active: form.gender_requirement===0 }" @click="form.gender_requirement=0">不限</view>
					<view class="radio" :class="{ active: form.gender_requirement===1 }" @click="form.gender_requirement=1">男</view>
					<view class="radio" :class="{ active: form.gender_requirement===2 }" @click="form.gender_requirement=2">女</view>
				</view>
			</view>
			<view class="field">
				<text class="label">结算方式 *</text>
				<view class="radios">
					<view v-for="(t, k) in settleTypes" :key="k" class="radio" :class="{ active: form.settlement_type===Number(k) }" @click="form.settlement_type=Number(k)">
						<text>{{ t }}</text>
					</view>
				</view>
			</view>
			<view class="field">
				<text class="label">工作时段(可选)</text>
				<view class="row">
					<input class="input small" v-model="form.work_time_start" placeholder="HH:MM" />
					<text class="dash">—</text>
					<input class="input small" v-model="form.work_time_end" placeholder="HH:MM" />
				</view>
			</view>
			<view class="field">
				<text class="label">岗位描述 *</text>
				<textarea class="textarea" v-model="form.description" maxlength="2000" placeholder="详细描述工作内容、要求、地点..."></textarea>
			</view>
			<view class="field">
				<text class="label">岗位要求(可选)</text>
				<textarea class="textarea" v-model="form.requirements" maxlength="2000" placeholder="学历、经验、技能要求..."></textarea>
			</view>

			<button class="btn-primary save" :disabled="saving" @click="onSubmit">
				{{ saving ? '提交中...' : '提交审核' }}
			</button>
			<view class="hint muted">提交后状态为"待审核",管理端通过后变成"招聘中"。</view>
		</view>
	</view>
</template>

<script>
	import { useAppStore } from '@/store/app'
	import { jobApi } from '@/api/job'
	import { SALARY_TYPE_TEXT, SETTLE_TYPE_TEXT } from '@/utils/constants'
	import { toastSuccess, toastError } from '@/utils/ui'

	export default {
		data() {
			return {
				form: this.emptyForm(),
				saving: false,
			}
		},
		computed: {
			categories() { return useAppStore().activeCategories },
			salaryTypes() { return SALARY_TYPE_TEXT },
			settleTypes() { return SETTLE_TYPE_TEXT },
		},
		onLoad() { useAppStore().loadCategories(true) },
		methods: {
			emptyForm() {
				return {
					title: '', category_id: 0,
					salary_type: 2, salary_min: 0, salary_max: 0,
					location: '', recruit_count: 1,
					gender_requirement: 0, settlement_type: 1,
					work_time_start: '', work_time_end: '',
					description: '', requirements: '',
				}
			},
			async onSubmit() {
				const f = this.form
				if (!f.title || f.title.length < 2) return toastError('请填写岗位名称')
				if (!f.category_id) return toastError('请选择分类')
				if (!f.salary_type) return toastError('请选择薪资方式')
				if (!f.salary_min) return toastError('请填写薪资')
				if (!f.location) return toastError('请填写工作地点')
				if (!f.recruit_count) return toastError('请填写招募人数')
				if (!f.description || f.description.length < 10) return toastError('岗位描述至少 10 个字')
				this.saving = true
				try {
					await jobApi.create({ ...f })
					toastSuccess('已提交,等待审核')
					setTimeout(() => uni.navigateBack(), 600)
				} catch (e) {} finally { this.saving = false }
			},
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.form { margin: $spacing-sm; }
	.field { padding: 20rpx 0; border-bottom: 2rpx solid $border-color-light; }
	.field:last-of-type { border-bottom: none; }
	.label { font-size: $font-sm; color: $text-secondary; }
	.input { display: block; margin-top: 8rpx; font-size: $font-lg; color: $text-primary; }
	.textarea { display: block; margin-top: 8rpx; font-size: $font-base; color: $text-primary; width: 100%; height: 200rpx; }
	.radios { display: flex; gap: $spacing-xs; flex-wrap: wrap; margin-top: 8rpx; }
	.radio {
		padding: 10rpx 24rpx; background: $bg-page; border-radius: $radius-pill; font-size: $font-sm; color: $text-regular;
		&.active { background: $brand-primary-bg; color: $brand-primary; }
	}
	.row { display: flex; align-items: center; gap: $spacing-sm; margin-top: 8rpx; }
	.input.small { width: 200rpx; padding: 12rpx 20rpx; background: $bg-page; border-radius: $radius-sm; }
	.dash { color: $text-secondary; }
	.unit { color: $text-secondary; font-size: $font-base; }
	.save { width: 100%; margin-top: $spacing-md; }
	.hint { text-align: center; margin-top: $spacing-sm; }
	.muted { color: $text-secondary; font-size: $font-xs; }
</style>
