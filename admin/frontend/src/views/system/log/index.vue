<template>
  <div class="app-container">
    <div class="search-form">
      <el-form :inline="true" @submit.prevent="reload">
        <el-form-item
          ><el-input v-model="keyword" placeholder="用户/路径/操作" clearable @clear="reload"
        /></el-form-item>
        <el-form-item><el-button type="primary" @click="reload">查询</el-button></el-form-item>
        <el-form-item v-permission="'log:delete'">
          <el-button type="danger" :disabled="!selection.length" @click="onBatchDelete"
            >批量删除 ({{ selection.length }})</el-button
          >
        </el-form-item>
      </el-form>
    </div>

    <el-table
      v-loading="loading"
      :data="list"
      border
      @selection-change="(rows) => (selection = rows)"
    >
      <el-table-column type="selection" width="48" />
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户" width="120" />
      <el-table-column prop="action" label="操作" width="240" />
      <el-table-column prop="method" label="方法" width="80" />
      <el-table-column prop="path" label="路径" />
      <el-table-column prop="ip" label="IP" width="140" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.response_status < 400 ? 'success' : 'danger'">{{
            row.response_status
          }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="latency_ms" label="耗时(ms)" width="100" />
      <el-table-column prop="created_at" label="时间" width="180" />
    </el-table>

    <Pagination
      v-model="page"
      :total="total"
      style="margin-top: 16px; justify-content: flex-end"
      @change="reload"
    />
  </div>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import { logApi, type OperationLog } from '@/api/log';
import Pagination from '@/components/Pagination.vue';

const list = ref<OperationLog[]>([]);
const total = ref(0);
const loading = ref(false);
const keyword = ref('');
const page = reactive({ page: 1, size: 20 });
const selection = ref<OperationLog[]>([]);

async function reload() {
  loading.value = true;
  try {
    const r = await logApi.list(page.page, page.size, keyword.value);
    list.value = r.list;
    total.value = r.total;
  } finally {
    loading.value = false;
  }
}

async function onBatchDelete() {
  await ElMessageBox.confirm(`确定删除选中的 ${selection.value.length} 条日志？`, '提示', {
    type: 'warning',
  });
  await logApi.batchRemove(selection.value.map((l) => l.id));
  ElMessage.success('已删除');
  reload();
}

onMounted(reload);
</script>
