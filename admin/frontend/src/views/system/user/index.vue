<template>
  <div class="app-container">
    <div class="search-form">
      <el-form :inline="true" @submit.prevent="reload">
        <el-form-item>
          <el-input
            v-model="keyword"
            placeholder="搜索用户名/昵称/邮箱"
            clearable
            @clear="reload"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="reload">查询</el-button>
          <el-button v-permission="'user:create'" type="success" @click="openCreate"
            >新增</el-button
          >
          <el-button
            v-permission="'user:delete'"
            type="danger"
            :disabled="!selection.length"
            @click="onBatchDelete"
          >
            批量删除 ({{ selection.length }})
          </el-button>
          <el-dropdown style="margin-left: 8px" @command="onExport">
            <el-button
              >导出 <el-icon><ArrowDown /></el-icon
            ></el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="excel">Excel</el-dropdown-item>
                <el-dropdown-item command="csv">CSV</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
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
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="nickname" label="昵称" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="phone" label="手机" />
      <el-table-column label="角色">
        <template #default="{ row }">
          <el-tag v-for="r in row.roles" :key="r.id" size="small" style="margin-right: 4px">{{
            r.name
          }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">{{
            row.status === 1 ? '启用' : '禁用'
          }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button v-permission="'user:update'" link type="primary" @click="openEdit(row as User)"
            >编辑</el-button
          >
          <el-button
            v-permission="'user:reset-password'"
            link
            type="warning"
            @click="resetPwd(row as User)"
            >重置密码</el-button
          >
          <el-button
            v-permission="'user:update'"
            link
            :type="(row as User).status === 1 ? 'info' : 'success'"
            @click="toggleStatus(row as User)"
          >
            {{ (row as User).status === 1 ? '禁用' : '启用' }}
          </el-button>
          <el-button v-permission="'user:delete'" link type="danger" @click="onDelete(row as User)"
            >删除</el-button
          >
        </template>
      </el-table-column>
    </el-table>

    <Pagination
      v-model="page"
      :total="total"
      style="margin-top: 16px; justify-content: flex-end"
      @change="reload"
    />

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑用户' : '新增用户'" width="520px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="!!editing" />
        </el-form-item>
        <el-form-item v-if="!editing" label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="昵称"><el-input v-model="form.nickname" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
        <el-form-item label="手机"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role_ids" multiple style="width: 100%">
            <el-option v-for="r in roleOptions" :key="r.id" :label="r.name" :value="r.id" />
          </el-select>
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
import { ArrowDown } from '@element-plus/icons-vue';
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import { type Role, roleApi } from '@/api/role';
import { type CreateUserReq, type UpdateUserReq, type User, userApi } from '@/api/user';
import Pagination from '@/components/Pagination.vue';

const list = ref<User[]>([]);
const total = ref(0);
const loading = ref(false);
const keyword = ref('');
const page = reactive({ page: 1, size: 10 });
const selection = ref<User[]>([]);

const dialogVisible = ref(false);
const editing = ref<User | null>(null);
const formRef = ref<FormInstance>();
const saving = ref(false);
const form = reactive<CreateUserReq & UpdateUserReq>({
  username: '',
  password: '',
  nickname: '',
  email: '',
  phone: '',
  role_ids: [],
});
const rules: FormRules = {
  username: [{ required: true, min: 3, max: 64, message: '3-64 字符', trigger: 'blur' }],
  password: [{ required: true, min: 6, message: '至少 6 位', trigger: 'blur' }],
};

const roleOptions = ref<Role[]>([]);

async function reload() {
  loading.value = true;
  try {
    const r = await userApi.list(page.page, page.size, keyword.value);
    list.value = r.list;
    total.value = r.total;
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editing.value = null;
  Object.assign(form, {
    username: '',
    password: '',
    nickname: '',
    email: '',
    phone: '',
    role_ids: [],
  });
  dialogVisible.value = true;
}

function openEdit(row: User) {
  editing.value = row;
  Object.assign(form, {
    username: row.username,
    nickname: row.nickname,
    email: row.email,
    phone: row.phone,
    role_ids: row.roles.map((r) => r.id),
  });
  dialogVisible.value = true;
}

async function onSubmit() {
  if (!formRef.value) return;
  await formRef.value.validate();
  saving.value = true;
  try {
    if (editing.value) {
      const { username: _, password: _p, ...rest } = form;
      await userApi.update(editing.value.id, rest);
    } else {
      await userApi.create({ ...form });
    }
    ElMessage.success('保存成功');
    dialogVisible.value = false;
    reload();
  } finally {
    saving.value = false;
  }
}

async function onDelete(row: User) {
  await ElMessageBox.confirm(`确定删除用户 ${row.username}？`, '提示', { type: 'warning' });
  await userApi.remove(row.id);
  ElMessage.success('已删除');
  reload();
}

async function onBatchDelete() {
  await ElMessageBox.confirm(`确定删除选中的 ${selection.value.length} 个用户？`, '提示', {
    type: 'warning',
  });
  await userApi.batchRemove(selection.value.map((u) => u.id));
  ElMessage.success('已删除');
  reload();
}

async function resetPwd(row: User) {
  const { value } = await ElMessageBox.prompt('输入新密码', '重置密码', {
    inputPattern: /.{6,}/,
    inputErrorMessage: '至少 6 位',
  });
  await userApi.resetPassword(row.id, value);
  ElMessage.success('密码已重置');
}

async function toggleStatus(row: User) {
  const next = row.status === 1 ? 2 : 1;
  await userApi.changeStatus(row.id, next as 1 | 2);
  ElMessage.success('状态已更新');
  reload();
}

async function onExport(cmd: 'excel' | 'csv') {
  const blob =
    cmd === 'excel'
      ? await userApi.exportExcel(keyword.value)
      : await userApi.exportCSV(keyword.value);
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `users_${Date.now()}.${cmd === 'excel' ? 'xlsx' : 'csv'}`;
  a.click();
  URL.revokeObjectURL(url);
}

onMounted(async () => {
  roleOptions.value = (await roleApi.list(1, 200)).list;
  reload();
});
</script>
