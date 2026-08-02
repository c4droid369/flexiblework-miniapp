<template>
  <div class="app-container">
    <div class="search-form">
      <el-form :inline="true" @submit.prevent="reload">
        <el-form-item>
          <el-input v-model="keyword" placeholder="搜索名称/编码" clearable @clear="reload" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="reload">查询</el-button>
          <el-button v-permission="'role:create'" type="success" @click="openCreate"
            >新增</el-button
          >
          <el-button
            v-permission="'role:delete'"
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
      <el-table-column prop="code" label="编码" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">{{
            row.status === 1 ? '启用' : '禁用'
          }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button v-permission="'role:update'" link type="primary" @click="openEdit(row as Role)"
            >编辑</el-button
          >
          <el-button
            v-permission="'role:assign-menu'"
            link
            type="warning"
            @click="openAssign(row as Role)"
            >分配菜单</el-button
          >
          <el-button v-permission="'role:delete'" link type="danger" @click="onDelete(row as Role)"
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

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑角色' : '新增角色'" width="520px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="编码">
          <el-input v-model="form.code" :disabled="!!editing" />
        </el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="描述"
          ><el-input v-model="form.description" type="textarea"
        /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSubmit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="assignVisible" title="分配菜单" width="420px">
      <el-tree
        ref="treeRef"
        :data="menuTree"
        :props="{ label: 'title', children: 'children' }"
        show-checkbox
        node-key="id"
        default-expand-all
      />
      <template #footer>
        <el-button @click="assignVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitAssign">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ArrowDown } from '@element-plus/icons-vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import type { MenuTree } from '@/api/auth';
import { menuApi } from '@/api/menu';
import { type CreateRoleReq, type Role, roleApi } from '@/api/role';
import Pagination from '@/components/Pagination.vue';

const list = ref<Role[]>([]);
const total = ref(0);
const loading = ref(false);
const keyword = ref('');
const page = reactive({ page: 1, size: 10 });
const selection = ref<Role[]>([]);

const dialogVisible = ref(false);
const editing = ref<Role | null>(null);
const form = reactive<CreateRoleReq>({
  code: '',
  name: '',
  description: '',
  sort: 0,
  menu_ids: [],
});
const saving = ref(false);

const assignVisible = ref(false);
const assignRole = ref<Role | null>(null);
const menuTree = ref<MenuTree[]>([]);
const treeRef = ref();

async function reload() {
  loading.value = true;
  try {
    const r = await roleApi.list(page.page, page.size, keyword.value);
    list.value = r.list;
    total.value = r.total;
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editing.value = null;
  Object.assign(form, { code: '', name: '', description: '', sort: 0, menu_ids: [] });
  dialogVisible.value = true;
}

function openEdit(row: Role) {
  editing.value = row;
  Object.assign(form, {
    code: row.code,
    name: row.name,
    description: row.description,
    sort: row.sort,
    menu_ids: row.menu_ids,
  });
  dialogVisible.value = true;
}

async function onSubmit() {
  saving.value = true;
  try {
    if (editing.value) {
      await roleApi.update(editing.value.id, form);
    } else {
      await roleApi.create({ ...form });
    }
    ElMessage.success('保存成功');
    dialogVisible.value = false;
    reload();
  } finally {
    saving.value = false;
  }
}

async function onDelete(row: Role) {
  await ElMessageBox.confirm(`确定删除角色 ${row.name}？`, '提示', { type: 'warning' });
  await roleApi.remove(row.id);
  ElMessage.success('已删除');
  reload();
}

async function onBatchDelete() {
  await ElMessageBox.confirm(`确定删除选中的 ${selection.value.length} 个角色？`, '提示', {
    type: 'warning',
  });
  await roleApi.batchRemove(selection.value.map((r) => r.id));
  ElMessage.success('已删除');
  reload();
}

async function openAssign(row: Role) {
  assignRole.value = row;
  menuTree.value = await menuApi.tree();
  assignVisible.value = true;
  await Promise.resolve();
  if (treeRef.value) {
    treeRef.value.setCheckedKeys(row.menu_ids, false);
  }
}

async function submitAssign() {
  if (!assignRole.value) return;
  const ids = treeRef.value.getCheckedKeys() as number[];
  await roleApi.assignMenus(assignRole.value.id, ids);
  ElMessage.success('已保存');
  assignVisible.value = false;
  reload();
}

async function onExport(cmd: 'excel' | 'csv') {
  const blob =
    cmd === 'excel'
      ? await roleApi.exportExcel(keyword.value)
      : await roleApi.exportCSV(keyword.value);
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `roles_${Date.now()}.${cmd === 'excel' ? 'xlsx' : 'csv'}`;
  a.click();
  URL.revokeObjectURL(url);
}

onMounted(reload);
</script>
