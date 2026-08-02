<template>
	<web-view :src="url"></web-view>
</template>

<script>
	export default {
		data() { return { url: '' } },
		onLoad(q) { this.url = q.url || '' },
	}
</script>
