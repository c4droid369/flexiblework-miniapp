<template>
  <div class="app-container">
    <div class="search-form">
      <el-tabs v-model="tab" @tab-change="reload">
        <el-tab-pane label="雇主资质" name="employer" />
        <el-tab-pane label="学生认证" name="student" />
      </el-tabs>
    </div>

    <!-- 雇主资质 -->
    <el-table v-if="tab==='employer'" v-loading="loading" :data="empList" border>
      <el-table-column prop="user_id" label="用户ID" width="100" />
      <el-table-column prop="username" label="账号" width="140" />
      <el-table-column prop="company_name" label="公司" min-width="180" />
      <el-table-column prop="contact_name" label="联系人" width="100" />
      <el-table-column prop="contact_phone" label="电话" width="140" />
      <el-table-column prop="business_license_no" label="营业执照号" width="200" show-overflow-tooltip />
      <el-table-column label="营业执照" width="100">
        <template #default="{ row }">
          <el-image
            v-if="row.business_license_img"
            :src="row.business_license_img"
            :preview-src-list="[row.business_license_img]"
            fit="cover"
            style="width: 60px; height: 60px; border-radius: 4px"
            preview-teleported
          />
          <span v-else class="muted">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="提交时间" width="180">
        <template #default="{ row }">{{ formatTs(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button v-permission="'cert:audit'" link type="success" @click="auditEmployer(row as EmployerCertItem, 2)">通过</el-button>
          <el-button v-permission="'cert:audit'" link type="danger" @click="auditEmployer(row as EmployerCertItem, 3)">拒绝</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 学生认证 -->
    <el-table v-else v-loading="loading" :data="stuList" border>
      <el-table-column prop="user_id" label="用户ID" width="100" />
      <el-table-column prop="username" label="账号" width="140" />
      <el-table-column prop="real_name" label="姓名" width="100" />
      <el-table-column prop="school" label="学校" width="160" />
      <el-table-column prop="college" label="学院" width="160" />
      <el-table-column prop="major" label="专业" width="160" />
      <el-table-column prop="student_no" label="学号" width="140" />
      <el-table-column label="证件" width="280">
        <template #default="{ row }">
          <el-image
            v-if="row.id_card_front"
            :src="row.id_card_front"
            :preview-src-list="[row.id_card_front, row.id_card_back, row.student_card].filter(Boolean) as string[]"
            fit="cover"
            style="width: 56px; height: 56px; border-radius: 4px; margin-right: 6px"
            preview-teleported
          />
          <el-image
            v-if="row.id_card_back"
            :src="row.id_card_back"
            :preview-src-list="[row.id_card_front, row.id_card_back, row.student_card].filter(Boolean) as string[]"
            fit="cover"
            style="width: 56px; height: 56px; border-radius: 4px; margin-right: 6px"
            preview-teleported
          />
          <el-image
            v-if="row.student_card"
            :src="row.student_card"
            :preview-src-list="[row.id_card_front, row.id_card_back, row.student_card].filter(Boolean) as string[]"
            fit="cover"
            style="width: 56px; height: 56px; border-radius: 4px"
            preview-teleported
          />
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="提交时间" width="180">
        <template #default="{ row }">{{ formatTs(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button v-permission="'cert:audit'" link type="success" @click="auditStudent(row as StudentCertItem, 2)">通过</el-button>
          <el-button v-permission="'cert:audit'" link type="danger" @click="auditStudent(row as StudentCertItem, 3)">拒绝</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Audit dialog with remark -->
    <el-dialog v-model="dialogVisible" :title="actionTitle" width="480px">
      <el-form label-width="84px">
        <el-form-item label="审核意见">
          <el-input v-model="auditRemark" type="textarea" :rows="4" maxlength="255" show-word-limit placeholder="可选,拒绝时建议填写原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button :type="actionType" :loading="saving" @click="confirmAudit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { certApi, type EmployerCertItem, type StudentCertItem } from '@/api/cert';

const tab = ref<'employer' | 'student'>('employer');
const empList = ref<EmployerCertItem[]>([]);
const stuList = ref<StudentCertItem[]>([]);
const loading = ref(false);

const dialogVisible = ref(false);
const saving = ref(false);
const auditRemark = ref('');
const pendingAction = ref<2 | 3>(2);
const pendingKind = ref<'employer' | 'student'>('employer');
const pendingUserId = ref<number>(0);

const actionTitle = computed(() => {
  const pass = pendingAction.value === 2;
  return (pass ? '通过' : '拒绝') + (pendingKind.value === 'employer' ? '雇主资质' : '学生认证');
});
const actionType = computed(() => (pendingAction.value === 2 ? 'primary' : 'danger'));

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
    if (tab.value === 'employer') {
      empList.value = await certApi.listPendingEmployer();
    } else {
      stuList.value = await certApi.listPendingStudent();
    }
  } finally {
    loading.value = false;
  }
}

function auditEmployer(row: EmployerCertItem, action: 2 | 3) {
  pendingKind.value = 'employer';
  pendingUserId.value = row.user_id;
  pendingAction.value = action;
  auditRemark.value = '';
  dialogVisible.value = true;
}

function auditStudent(row: StudentCertItem, action: 2 | 3) {
  pendingKind.value = 'student';
  pendingUserId.value = row.user_id;
  pendingAction.value = action;
  auditRemark.value = '';
  dialogVisible.value = true;
}

async function confirmAudit() {
  saving.value = true;
  try {
    if (pendingKind.value === 'employer') {
      await certApi.auditEmployer(pendingUserId.value, { action: pendingAction.value, remark: auditRemark.value });
    } else {
      await certApi.auditStudent(pendingUserId.value, { action: pendingAction.value, remark: auditRemark.value });
    }
    ElMessage.success(pendingAction.value === 2 ? '已通过' : '已拒绝');
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
.muted { color: #909399; }
</style>
