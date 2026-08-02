<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2 class="title">Admin Template</h2>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        @submit.prevent="onSubmit"
      >
        <el-form-item label="Username" prop="username">
          <el-input v-model="form.username" placeholder="admin" autocomplete="username" />
        </el-form-item>
        <el-form-item label="Password" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="••••••"
            show-password
            autocomplete="current-password"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-button type="primary" :loading="loading" style="width: 100%" @click="onSubmit"
          >登录</el-button
        >
      </el-form>
      <p class="hint">默认账号 admin / admin123 · 请在生产环境修改</p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const formRef = ref<FormInstance>();
const loading = ref(false);
const form = reactive({ username: '', password: '' });
const rules: FormRules = {
  username: [{ required: true, min: 3, message: '至少 3 个字符', trigger: 'blur' }],
  password: [{ required: true, min: 6, message: '至少 6 个字符', trigger: 'blur' }],
};

async function onSubmit() {
  if (!formRef.value) return;
  try {
    await formRef.value.validate();
  } catch {
    return;
  }
  loading.value = true;
  try {
    await auth.login(form.username, form.password);
    ElMessage.success('登录成功');
    const redirect = (route.query.redirect as string) || '/';
    router.push(redirect);
  } catch {
    // request interceptor surfaces error message
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1f4e9c, #0a1f44);
}
.login-card {
  width: 360px;
  padding: 16px 8px;
}
.title {
  text-align: center;
  margin: 0 0 24px;
  color: var(--app-primary);
}
.hint {
  color: var(--app-text-secondary);
  font-size: 12px;
  text-align: center;
  margin-top: 16px;
}
</style>
