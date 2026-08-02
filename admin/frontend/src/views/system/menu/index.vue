<template>
  <div class="app-container">
    <div class="search-form">
      <el-button v-permission="'menu:create'" type="primary" @click="openCreate(0)"
        >新增根菜单</el-button
      >
      <el-button @click="reload">刷新</el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="menuTree"
      row-key="id"
      :tree-props="{ children: 'children' }"
      border
      default-expand-all
    >
      <el-table-column prop="title" label="名称" />
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="typeTag(row.type)">{{ typeLabel(row.type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="path" label="路径" />
      <el-table-column prop="component" label="组件" />
      <el-table-column prop="perm_code" label="权限码" />
      <el-table-column prop="icon" label="图标" />
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button
            v-permission="'menu:create'"
            link
            type="primary"
            @click="openCreate((row as MenuTree).id)"
            >新增子项</el-button
          >
          <el-button
            v-permission="'menu:update'"
            link
            type="warning"
            @click="openEdit(row as MenuTree)"
            >编辑</el-button
          >
          <el-button
            v-permission="'menu:delete'"
            link
            type="danger"
            @click="onDelete(row as MenuTree)"
            >删除</el-button
          >
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑菜单' : '新增菜单'" width="540px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="父级">
          <el-tree-select
            v-model="form.parent_id"
            :data="parentOptions"
            node-key="id"
            :props="treeProps"
            check-strictly
            clearable
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="类型">
          <el-radio-group v-model="form.type">
            <el-radio :value="1">目录</el-radio>
            <el-radio :value="2">菜单</el-radio>
            <el-radio :value="3">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item v-if="form.type !== 3" label="路径"
          ><el-input v-model="form.path"
        /></el-form-item>
        <el-form-item v-if="form.type === 2" label="组件"
          ><el-input v-model="form.component" placeholder="views/system/user/index.vue"
        /></el-form-item>
        <el-form-item label="权限码"
          ><el-input v-model="form.perm_code" placeholder="resource:action"
        /></el-form-item>
        <el-form-item label="图标"><el-input v-model="form.icon" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item v-if="form.type !== 3" label="可见">
          <el-switch v-model="form.visible" />
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
import { ElMessage, ElMessageBox } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';
import type { MenuTree } from '@/api/auth';
import { type CreateMenuReq, menuApi } from '@/api/menu';

const menuTree = ref<MenuTree[]>([]);
const loading = ref(false);

const dialogVisible = ref(false);
const editing = ref<MenuTree | null>(null);
const form = reactive<CreateMenuReq>({
  parent_id: 0,
  type: 2,
  name: '',
  title: '',
  path: '',
  component: '',
  perm_code: '',
  icon: '',
  sort: 0,
  visible: true,
});
const saving = ref(false);

const parentOptions = computed(() => [{ id: 0, title: '— 根级 —', children: menuTree.value }]);

// el-tree-select uses node-key for value binding; we only need label + children.
// Declared as a typed const so TS doesn't fight the literal against the
// generic TreeOptionProps type.
const treeProps = { label: 'title', children: 'children' } as const;

async function reload() {
  loading.value = true;
  try {
    menuTree.value = await menuApi.tree();
  } finally {
    loading.value = false;
  }
}

function openCreate(parentId: number) {
  editing.value = null;
  Object.assign(form, {
    parent_id: parentId,
    type: 2,
    name: '',
    title: '',
    path: '',
    component: '',
    perm_code: '',
    icon: '',
    sort: 0,
    visible: true,
  });
  dialogVisible.value = true;
}

function openEdit(row: MenuTree) {
  editing.value = row;
  Object.assign(form, {
    parent_id: row.parent_id,
    type: row.type,
    name: row.name,
    title: row.title,
    path: row.path,
    component: row.component,
    perm_code: row.perm_code,
    icon: row.icon,
    sort: row.sort,
    visible: row.visible,
  });
  dialogVisible.value = true;
}

async function onSubmit() {
  saving.value = true;
  try {
    if (editing.value) {
      await menuApi.update(editing.value.id, form);
    } else {
      await menuApi.create(form);
    }
    ElMessage.success('保存成功');
    dialogVisible.value = false;
    reload();
  } finally {
    saving.value = false;
  }
}

async function onDelete(row: MenuTree) {
  await ElMessageBox.confirm(`确定删除菜单 ${row.title}？（子项一并删除）`, '提示', {
    type: 'warning',
  });
  await menuApi.remove(row.id);
  ElMessage.success('已删除');
  reload();
}

function typeLabel(t: number) {
  return { 1: '目录', 2: '菜单', 3: '按钮' }[t] ?? '-';
}
function typeTag(t: number) {
  return ({ 1: 'info', 2: 'primary', 3: 'warning' } as const)[t as 1 | 2 | 3] ?? 'info';
}

onMounted(reload);
</script>
