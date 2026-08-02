<template>
  <div class="app-container">
    <div class="search-form">
      <el-form :inline="true" @submit.prevent="reload">
        <el-form-item>
          <el-button @click="reload">刷新</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-table v-loading="loading" :data="list" border>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="岗位名称" min-width="180" />
      <el-table-column prop="category_name" label="分类" width="120" />
      <el-table-column prop="employer_name" label="发布方" width="160" />
      <el-table-column label="薪资" width="140">
        <template #default="{ row }">{{ row.salary_text }}</template>
      </el-table-column>
      <el-table-column prop="location" label="地点" width="160" show-overflow-tooltip />
      <el-table-column prop="recruit_count" label="招人数" width="80" />
      <el-table-column prop="created_at" label="提交时间" width="180">
        <template #default="{ row }">{{ formatTs(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row as AdminJob)">查看</el-button>
          <el-button v-permission="'job:audit'" link type="success" @click="onAudit(row as AdminJob, 2)">通过</el-button>
          <el-button v-permission="'job:audit'" link type="danger" @click="onAudit(row as AdminJob, 4)">拒绝</el-button>
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

    <!-- Detail / audit dialog -->
    <el-dialog v-model="dialogVisible" :title="`岗位审核 — ${current?.title ?? ''}`" width="720px">
      <div v-if="current" class="detail">
        <div class="row"><span class="lbl">岗位</span><span class="val">{{ current.title }}</span></div>
        <div class="row"><span class="lbl">发布方</span><span class="val">{{ current.employer_name }}</span></div>
        <div class="row"><span class="lbl">分类</span><span class="val">{{ current.category_name }}</span></div>
        <div class="row"><span class="lbl">薪资</span><span class="val">{{ current.salary_text }}</span></div>
        <div class="row"><span class="lbl">地点</span><span class="val">{{ current.location }}</span></div>
        <div class="row"><span class="lbl">招人数</span><span class="val">{{ current.recruit_count }}</span></div>
        <el-divider />
        <div class="row col"><span class="lbl">岗位描述</span><div class="val pre">{{ current.description }}</div></div>
        <div class="row col" v-if="current.requirements"><span class="lbl">岗位要求</span><div class="val pre">{{ current.requirements }}</div></div>

        <el-form label-width="84px" style="margin-top: 12px">
          <el-form-item label="审核意见">
            <el-input v-model="auditRemark" type="textarea" :rows="3" maxlength="255" show-word-limit />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">关闭</el-button>
        <el-button type="danger" :loading="saving" @click="onAudit(current!, 4)">拒绝</el-button>
        <el-button type="primary" :loading="saving" @click="onAudit(current!, 2)">通过</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElMessage, type FormInstance } from 'element-plus';
import { jobApi, type AdminJob } from '@/api/job';

const list = ref<AdminJob[]>([]);
const total = ref(0);
const page = ref(1);
const size = ref(20);
const loading = ref(false);
const saving = ref(false);

const dialogVisible = ref(false);
const current = ref<AdminJob | null>(null);
const auditRemark = ref('');

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
    const r = await jobApi.pending(page.value, size.value);
    list.value = r.list;
    total.value = r.total;
  } finally {
    loading.value = false;
  }
}

function openDetail(row: AdminJob) {
  current.value = row;
  auditRemark.value = '';
  dialogVisible.value = true;
}

async function onAudit(row: AdminJob, action: 2 | 4) {
  saving.value = true;
  try {
    await jobApi.audit(row.id, { action, remark: auditRemark.value });
    ElMessage.success(action === 2 ? '已通过' : '已拒绝');
    dialogVisible.value = false;
    await reload();
  } finally {
    saving.value = false;
  }
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
.val.pre { white-space: pre-wrap; line-height: 1.7; }
</style>
