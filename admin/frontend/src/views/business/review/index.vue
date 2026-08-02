<template>
  <div class="app-container">
    <div class="search-form">
      <el-alert
        type="info"
        :closable="false"
        title="提示"
        description="当前为评价管理概览。如需查看某订单的全部评价,请进入「订单监控」→ 订单详情。删除评价会同时影响双方历史评分。"
      />
    </div>

    <el-table v-loading="loading" :data="list" border>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="order_id" label="订单ID" width="100" />
      <el-table-column prop="from_name" label="评价人" width="160" />
      <el-table-column label="方向" width="100">
        <template #default="{ row }">
          <el-tag :type="row.role === 1 ? 'primary' : 'success'">
            {{ row.role === 1 ? '学生→雇主' : '雇主→学生' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="评分" width="120">
        <template #default="{ row }">
          <el-rate v-model="row.rating" disabled show-score :max="5" />
        </template>
      </el-table-column>
      <el-table-column prop="content" label="内容" min-width="220" show-overflow-tooltip />
      <el-table-column label="标签" min-width="200">
        <template #default="{ row }">
          <el-tag v-for="t in (row.tags || [])" :key="t" size="small" style="margin-right: 4px">{{ t }}</el-tag>
          <span v-if="!row.tags || !row.tags.length" class="muted">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="提交时间" width="180">
        <template #default="{ row }">{{ formatTs(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button v-permission="'review:delete'" link type="danger" @click="onDelete(row as AdminReview)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="size"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="reload"
        @size-change="reload"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { reviewApi, type AdminReview } from '@/api/review';

const list = ref<AdminReview[]>([]);
const total = ref(0);
const page = ref(1);
const size = ref(20);
const loading = ref(false);

function formatTs(s: string): string {
  if (!s) return '—';
  const d = new Date(s);
  if (isNaN(d.getTime())) return s;
  const pad = (n: number) => (n < 10 ? '0' + n : '' + n);
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

async function reload() {
  loading.value = true;
  try {
    const r = await reviewApi.list(page.value, size.value);
    list.value = r.list || [];
    total.value = r.total || 0;
  } finally {
    loading.value = false;
  }
}

async function onDelete(row: AdminReview) {
  await ElMessageBox.confirm(`确认删除评价 #${row.id}?`, '提示', { type: 'warning' });
  await reviewApi.remove(row.id);
  ElMessage.success('已删除');
  await reload();
}

onMounted(reload);
</script>

<style scoped>
.app-container { padding: 16px; }
.pager { margin-top: 16px; text-align: right; }
.muted { color: #909399; }
</style>
