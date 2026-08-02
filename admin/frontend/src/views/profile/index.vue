<template>
  <div class="app-container">
    <el-card header="个人中心">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="用户名">{{ auth.profile?.username }}</el-descriptions-item>
        <el-descriptions-item label="昵称">{{ auth.profile?.nickname }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ auth.profile?.email || '—' }}</el-descriptions-item>
        <el-descriptions-item label="手机">{{ auth.profile?.phone || '—' }}</el-descriptions-item>
        <el-descriptions-item label="角色">
          <el-tag v-for="r in auth.profile?.roles" :key="r.id" style="margin-right: 4px">{{
            r.name
          }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="最近登录">{{
          auth.profile?.last_login_at || '—'
        }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card style="margin-top: 16px" header="修改密码">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        style="max-width: 480px"
      >
        <el-form-item label="新密码" prop="password">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm">
          <el-input v-model="form.confirm" type="password" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="onSubmit">提交</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();
const formRef = ref<FormInstance>();
const loading = ref(false);
const form = reactive({ password: '', confirm: '' });
const rules: FormRules = {
  password: [{ required: true, min: 6, message: '至少 6 位', trigger: 'blur' }],
  confirm: [
    { required: true, message: '请再次输入', trigger: 'blur' },
    {
      validator: (_r, v, cb) => (v === form.password ? cb() : cb(new Error('两次密码不一致'))),
      trigger: 'blur',
    },
  ],
};

async function onSubmit() {
  if (!formRef.value) return;
  await formRef.value.validate();
  ElMessage.success('密码已更新（demo）');
  form.password = '';
  form.confirm = '';
}
</script>
