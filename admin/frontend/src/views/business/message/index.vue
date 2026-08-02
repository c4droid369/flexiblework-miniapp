<template>
  <div class="app-container">
    <el-card class="form-card">
      <template #header>
        <div class="card-header">
          <span>广播系统消息</span>
          <span class="muted">向指定用户群发送一条站内通知,所有目标用户的消息中心会立即收到。</span>
        </div>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="消息类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio :value="1">系统通知</el-radio>
            <el-radio :value="2">岗位相关</el-radio>
            <el-radio :value="3">订单相关</el-radio>
            <el-radio :value="4">评价相关</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="接收对象" prop="user_type">
          <el-radio-group v-model="form.user_type">
            <el-radio value="all">全部用户</el-radio>
            <el-radio value="student">仅学生</el-radio>
            <el-radio value="employer">仅雇主</el-radio>
            <el-radio value="admin">仅管理员</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" maxlength="128" show-word-limit placeholder="如:系统升级通知 / 周末活动" />
        </el-form-item>
        <el-form-item label="正文" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="6" maxlength="1000" show-word-limit placeholder="消息正文,会推送到目标用户的消息中心" />
        </el-form-item>
        <el-form-item label="跳转链接" prop="link">
          <el-input v-model="form.link" maxlength="255" placeholder="可选,小程序路径如 pages/jobs/jobs" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="sending" @click="onSubmit">发送广播</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="result-card" v-if="lastSent">
      <template #header><span>最近一次发送结果</span></template>
      <el-result icon="success" :title="`成功推送给 ${lastSent.count} 个用户`" sub-title="目标用户的消息中心会立即显示该通知">
        <template #extra>
          <el-button @click="onReset">继续发送</el-button>
        </template>
      </el-result>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { messageApi, type BroadcastReq } from '@/api/message';

const formRef = ref<FormInstance>();
const sending = ref(false);
const lastSent = ref<{ count: number; req: BroadcastReq } | null>(null);

const form = reactive<BroadcastReq>({
  type: 1,
  user_type: 'all',
  title: '',
  content: '',
  link: '',
});

const rules: FormRules = {
  type: [{ required: true, message: '请选择消息类型', trigger: 'change' }],
  user_type: [{ required: true, message: '请选择接收对象', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入正文', trigger: 'blur' }],
};

async function onSubmit() {
  if (!formRef.value) return;
  await formRef.value.validate();
  sending.value = true;
  try {
    const count = await messageApi.broadcast({ ...form });
    ElMessage.success(`已发送给 ${count} 个用户`);
    lastSent.value = { count, req: { ...form } };
  } finally {
    sending.value = false;
  }
}

function onReset() {
  form.type = 1;
  form.user_type = 'all';
  form.title = '';
  form.content = '';
  form.link = '';
  lastSent.value = null;
  formRef.value?.clearValidate();
}
</script>

<style scoped>
.app-container { padding: 16px; max-width: 760px; }
.form-card { margin-bottom: 16px; }
.card-header { display: flex; align-items: baseline; justify-content: space-between; }
.muted { color: #909399; font-size: 12px; }
.result-card { margin-top: 16px; }
</style>
