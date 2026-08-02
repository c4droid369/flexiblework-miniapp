<template>
  <div class="app-container">
    <el-row :gutter="16">
      <el-col v-for="card in cards" :key="card.title" :xs="24" :sm="12" :md="6">
        <el-card shadow="hover">
          <div class="stat">
            <el-icon :size="40" :color="card.color"><component :is="card.icon" /></el-icon>
            <div>
              <div class="num">{{ card.value }}</div>
              <div class="lbl">{{ card.title }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top: 16px" header="欢迎">
      <p style="margin: 0; line-height: 1.6">
        这是一个通用的 RBAC 后台管理模板。当前登录用户：
        <strong>{{ auth.profile?.nickname || auth.profile?.username }}</strong
        >， 角色：
        <el-tag v-for="r in auth.profile?.roles" :key="r.id" style="margin-right: 4px">{{
          r.name
        }}</el-tag>
      </p>
      <p style="margin-top: 12px">
        权限数：{{ auth.profile?.permissions.length ?? 0 }} · 可访问菜单数：{{
          auth.profile?.menus.length ?? 0
        }}
      </p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { DataAnalysis, Key, User, UserFilled } from '@element-plus/icons-vue';
import { computed } from 'vue';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();

const cards = computed(() => [
  { title: '用户', value: '-', icon: User, color: '#409eff' },
  { title: '角色', value: '-', icon: UserFilled, color: '#67c23a' },
  { title: '权限', value: auth.profile?.permissions.length ?? 0, icon: Key, color: '#e6a23c' },
  { title: '菜单', value: auth.profile?.menus.length ?? 0, icon: DataAnalysis, color: '#f56c6c' },
]);
</script>

<style scoped>
.stat {
  display: flex;
  align-items: center;
  gap: 16px;
}
.num {
  font-size: 24px;
  font-weight: 600;
}
.lbl {
  color: var(--app-text-secondary);
  font-size: 13px;
}
</style>
