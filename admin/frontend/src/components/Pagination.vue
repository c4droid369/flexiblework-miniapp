<template>
  <el-pagination
    v-model:current-page="page"
    v-model:page-size="size"
    :total="total"
    :page-sizes="[10, 20, 50, 100]"
    layout="total, sizes, prev, pager, next, jumper"
    background
    @current-change="$emit('change', { page, size })"
    @size-change="onSize"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{ modelValue: { page: number; size: number }; total: number }>();
const emit = defineEmits<{
  (e: 'update:modelValue', v: { page: number; size: number }): void;
  (e: 'change', v: { page: number; size: number }): void;
}>();

const page = computed({
  get: () => props.modelValue.page,
  set: (v) => emit('update:modelValue', { page: v, size: size.value }),
});
const size = computed({
  get: () => props.modelValue.size,
  set: (v) => emit('update:modelValue', { page: page.value, size: v }),
});
function onSize(v: number) {
  emit('update:modelValue', { page: 1, size: v });
  emit('change', { page: 1, size: v });
}
</script>
