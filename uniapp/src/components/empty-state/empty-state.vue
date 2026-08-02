<template>
	<view class="empty">
		<view class="emoji">{{ emoji }}</view>
		<text>{{ text }}</text>
		<view v-if="actionText" class="action" @click="$emit('action')">
			<text>{{ actionText }}</text>
		</view>
	</view>
</template>

<script>
	export default {
		name: 'EmptyState',
		props: {
			text: { type: String, default: '暂无数据' },
			emoji: { type: String, default: '📭' },
			actionText: { type: String, default: '' },
		},
	}
</script>

<style lang="scss" scoped>
	@import '@/uni.scss';
	.empty { display: flex; flex-direction: column; align-items: center; padding: $spacing-xl 0; color: $text-placeholder; }
	.emoji { font-size: 96rpx; margin-bottom: $spacing-sm; }
	.action { margin-top: $spacing-md; padding: 16rpx 40rpx; background: $brand-primary; color: #fff; border-radius: $radius-pill; font-size: $font-sm; }
</style>
