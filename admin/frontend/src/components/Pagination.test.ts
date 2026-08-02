import { mount } from '@vue/test-utils';
import { describe, expect, it, vi } from 'vitest';
import Pagination from '@/components/Pagination.vue';

// Element Plus's el-pagination needs the full app context to render. For
// these tests we only care about the wrapper's event flow, so stub it.
const ElPaginationStub = {
  name: 'ElPagination',
  props: ['modelValue', 'total', 'pageSizes', 'layout', 'background'],
  emits: ['update:modelValue', 'current-change', 'size-change'],
  template: '<div class="el-pagination-stub" />',
};

describe('Pagination.vue', () => {
  it('forwards current-change as a change event', async () => {
    const w = mount(Pagination, {
      props: { modelValue: { page: 2, size: 10 }, total: 100 },
      global: { stubs: { 'el-pagination': ElPaginationStub } },
    });

    // el-pagination emits `current-change` BEFORE its v-model updates the
    // parent — so the wrapper sees the stale `page` value. The contract is
    // "a change was reported", and consumers re-read modelValue after.
    await w.findComponent(ElPaginationStub).vm.$emit('current-change', 5);
    const updates = w.emitted('change');
    expect(updates).toBeTruthy();
    expect(updates![0]).toEqual([{ page: 2, size: 10 }]);
  });

  it('resets page to 1 on size-change', async () => {
    const w = mount(Pagination, {
      props: { modelValue: { page: 3, size: 10 }, total: 100 },
      global: { stubs: { 'el-pagination': ElPaginationStub } },
    });
    await w.findComponent(ElPaginationStub).vm.$emit('size-change', 20);
    const updates = w.emitted('change');
    expect(updates).toBeTruthy();
    expect(updates![0]).toEqual([{ page: 1, size: 20 }]);
  });
});

// guard
vi;
