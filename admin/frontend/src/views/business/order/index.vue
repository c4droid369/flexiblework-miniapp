<template>
  <div class="app-container">
    <div class="search-form">
      <el-form :inline="true" @submit.prevent="reload">
        <el-form-item>
          <el-select v-model="filterStatus" placeholder="订单状态" clearable style="width: 160px" @change="reload">
            <el-option v-for="(text, val) in ORDER_STATUS_TEXT" :key="val" :label="text" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button @click="reload">刷新</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-table v-loading="loading" :data="list" border>
      <el-table-column prop="order_no" label="订单号" width="200" />
      <el-table-column prop="job_title" label="岗位" min-width="180" show-overflow-tooltip />
      <el-table-column prop="employer_name" label="雇主" width="140" />
      <el-table-column prop="student_name" label="学生" width="120" />
      <el-table-column label="金额" width="100">
        <template #default="{ row }">¥{{ row.amount.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="ORDER_STATUS_TAG[row.status] || 'info'">
            {{ ORDER_STATUS_TEXT[row.status] || '—' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="支付方式" width="120">
        <template #default="{ row }">{{ row.pay_method || '—' }}</template>
      </el-table-column>
      <el-table-column prop="created_at" label="下单时间" width="180">
        <template #default="{ row }">{{ formatTs(row.created_at) }}</template>
      </el-table-column>
      <el-table-column prop="settled_at" label="结算时间" width="180">
        <template #default="{ row }">{{ row.settled_at ? formatTs(row.settled_at) : '—' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row as AdminOrder)">详情</el-button>
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

    <!-- Detail dialog -->
    <el-dialog v-model="dialogVisible" :title="`订单详情 — ${current?.order_no ?? ''}`" width="640px">
      <div v-if="current" class="detail">
        <div class="row"><span class="lbl">订单号</span><span class="val">{{ current.order_no }}</span></div>
        <div class="row"><span class="lbl">岗位</span><span class="val">{{ current.job_title }}</span></div>
        <div class="row"><span class="lbl">雇主</span><span class="val">{{ current.employer_name }}</span></div>
        <div class="row"><span class="lbl">学生</span><span class="val">{{ current.student_name }}</span></div>
        <div class="row"><span class="lbl">金额</span><span class="val">¥{{ current.amount.toFixed(2) }}</span></div>
        <div class="row"><span class="lbl">状态</span>
          <span class="val">
            <el-tag :type="ORDER_STATUS_TAG[current.status] || 'info'">
              {{ ORDER_STATUS_TEXT[current.status] || '—' }}
            </el-tag>
          </span>
        </div>
        <el-divider />
        <div class="row"><span class="lbl">下单</span><span class="val">{{ formatTs(current.created_at) }}</span></div>
        <div class="row" v-if="current.paid_at"><span class="lbl">支付</span><span class="val">{{ formatTs(current.paid_at) }}</span></div>
        <div class="row" v-if="current.started_at"><span class="lbl">开始</span><span class="val">{{ formatTs(current.started_at) }}</span></div>
        <div class="row" v-if="current.completed_at"><span class="lbl">完成</span><span class="val">{{ formatTs(current.completed_at) }}</span></div>
        <div class="row" v-if="current.confirmed_at"><span class="lbl">确认</span><span class="val">{{ formatTs(current.confirmed_at) }}</span></div>
        <div class="row" v-if="current.settled_at"><span class="lbl">结算</span><span class="val">{{ formatTs(current.settled_at) }}</span></div>
        <div v-if="current.cancel_reason" class="row"><span class="lbl">取消原因</span><span class="val">{{ current.cancel_reason }}</span></div>
        <div v-if="current.work_proof && current.work_proof.length" class="row col">
          <span class="lbl">上岗凭证</span>
          <div class="val proof-grid">
            <el-image
              v-for="(p, i) in current.work_proof"
              :key="i"
              :src="p"
              :preview-src-list="current.work_proof"
              fit="cover"
              style="width: 100px; height: 100px; border-radius: 4px"
              preview-teleported
            />
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { orderApi, ORDER_STATUS_TEXT, ORDER_STATUS_TAG, type AdminOrder } from '@/api/order';

const list = ref<AdminOrder[]>([]);
const total = ref(0);
const page = ref(1);
const size = ref(20);
const loading = ref(false);
const filterStatus = ref<number | undefined>(undefined);

const dialogVisible = ref(false);
const current = ref<AdminOrder | null>(null);

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
    const r = await orderApi.list(page.value, size.value, filterStatus.value);
    list.value = r.list;
    total.value = r.total;
  } finally {
    loading.value = false;
  }
}

function openDetail(row: AdminOrder) {
  current.value = row;
  dialogVisible.value = true;
}

onMounted(reload);
</script>

<style scoped>
.app-container { padding: 16px; }
.pager { margin-top: 16px; text-align: right; }
.detail .row { display: flex; padding: 6px 0; }
.detail .row.col { flex-direction: column; }
.lbl { width: 84px; color: #909399; }
.val { flex: 1; color: #303133; }
.proof-grid { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 6px; }
</style>
