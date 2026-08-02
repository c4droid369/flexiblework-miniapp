<template>
  <div class="app-container">
    <div class="search-form">
      <el-form :inline="true" @submit.prevent="reload">
        <el-form-item>
          <el-select v-model="filterStatus" placeholder="状态" clearable style="width: 140px" @change="reload">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button v-permission="'category:create'" type="success" @click="openCreate">新增分类</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-table v-loading="loading" :data="list" border>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="名称" min-width="140" />
      <el-table-column prop="icon" label="图标" width="160">
        <template #default="{ row }">
          <code class="muted">{{ row.icon || '—' }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="80" sortable />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">{{ formatTs(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button v-permission="'category:update'" link type="primary" @click="openEdit(row as Category)">编辑</el-button>
          <el-button v-permission="'category:delete'" link type="danger" @click="onDelete(row as Category)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Create / Edit dialog -->
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑分类' : '新增分类'" width="520px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="84px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" maxlength="255" placeholder="icon 名称或 URL" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="2">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" maxlength="255" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import { categoryApi, type Category, type CreateCategoryReq, type UpdateCategoryReq } from '@/api/category';

const list = ref<Category[]>([]);
const loading = ref(false);
const saving = ref(false);
const filterStatus = ref<1 | 2 | undefined>(undefined);

const dialogVisible = ref(false);
const editing = ref<Category | null>(null);
const formRef = ref<FormInstance>();
const form = ref<CreateCategoryReq & UpdateCategoryReq>({
  name: '',
  icon: '',
  sort: 0,
  status: 1,
  description: '',
});
const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
};

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
    list.value = await categoryApi.list(filterStatus.value);
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editing.value = null;
  form.value = { name: '', icon: '', sort: 0, status: 1, description: '' };
  dialogVisible.value = true;
}

function openEdit(row: Category) {
  editing.value = row;
  form.value = {
    name: row.name,
    icon: row.icon,
    sort: row.sort,
    status: row.status,
    description: row.description,
  };
  dialogVisible.value = true;
}

async function onSubmit() {
  if (!formRef.value) return;
  await formRef.value.validate();
  saving.value = true;
  try {
    if (editing.value) {
      await categoryApi.update(editing.value.id, form.value);
      ElMessage.success('已保存');
    } else {
      await categoryApi.create(form.value);
      ElMessage.success('已创建');
    }
    dialogVisible.value = false;
    await reload();
  } finally {
    saving.value = false;
  }
}

async function onDelete(row: Category) {
  await ElMessageBox.confirm(`确认删除分类「${row.name}」?`, '提示', { type: 'warning' });
  await categoryApi.remove(row.id);
  ElMessage.success('已删除');
  await reload();
}

onMounted(reload);
</script>

<style scoped>
.app-container { padding: 16px; }
.muted { color: #909399; font-size: 12px; }
</style>
