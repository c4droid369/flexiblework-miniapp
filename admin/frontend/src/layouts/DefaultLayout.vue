<template>
  <el-container class="layout">
    <el-aside :width="app.sidebarCollapsed ? '64px' : '220px'" class="aside">
      <div class="logo">
        <span v-if="!app.sidebarCollapsed">Admin Template</span>
        <span v-else>AT</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="app.sidebarCollapsed"
        :collapse-transition="false"
        background-color="#001529"
        text-color="#ffffffcc"
        active-text-color="#ffffff"
        router
      >
        <!-- Static entries always present. -->
        <el-menu-item index="/dashboard">
          <el-icon><DataLine /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>
        <el-menu-item index="/profile">
          <el-icon><User /></el-icon>
          <template #title>个人中心</template>
        </el-menu-item>

        <!-- Dynamic entries rendered from /auth/me menu tree. -->
        <template v-for="item in menuItems" :key="item.path">
          <el-sub-menu v-if="item.children?.length" :index="item.path">
            <template #title>
              <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
              <span>{{ item.title }}</span>
            </template>
            <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path">
              <el-icon v-if="child.icon"><component :is="child.icon" /></el-icon>
              <template #title>{{ child.title }}</template>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="item.path">
            <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <el-button :icon="app.sidebarCollapsed ? Expand : Fold" text @click="app.toggleSidebar" />
        <div class="spacer" />
        <el-dropdown @command="onCommand">
          <span class="user">
            <el-avatar :size="28">{{ auth.profile?.nickname?.charAt(0) ?? '?' }}</el-avatar>
            <span style="margin-left: 8px">{{
              auth.profile?.nickname || auth.profile?.username
            }}</span>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人中心</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>

      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { DataLine, Expand, Fold, User } from '@element-plus/icons-vue';
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import type { MenuTree } from '@/api/auth';
import { useAppStore } from '@/stores/app';
import { useAuthStore } from '@/stores/auth';

const app = useAppStore();
const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const activeMenu = computed(() => route.path);

interface SidebarItem {
  path: string;
  title: string;
  icon?: string;
  children?: SidebarItem[];
}

// Sidebar is driven by /auth/me's menu tree, NOT by router.options.routes.
// The latter is a static snapshot; dynamic routes registered via
// router.addRoute live in the runtime table only. Using the menus tree here
// also keeps the directory/grouping structure intact (seed emits System as
// a parent of User/Role/Menu/Log).

function toItem(m: MenuTree): SidebarItem | null {
  // Sidebar only renders directories (1) and routable menus (2). Buttons (3)
  // are surfaced via v-permission, not the sidebar.
  if (m.type !== 1 && m.type !== 2) return null;

  // Normalize path: must be a non-empty absolute URL path. Element Plus el-menu
  // in router mode rejects bare relative paths like "dashboard" with
  // "Invalid path" — vue-router cannot resolve them.
  const raw = (m.path ?? '').trim();
  if (!raw) return null;
  const path = raw.startsWith('/') ? raw : `/${raw}`;

  const children = m.children?.map(toItem).filter((c): c is SidebarItem => c !== null);

  return {
    path,
    title: m.title || m.name,
    icon: m.icon || undefined,
    children: children && children.length > 0 ? children : undefined,
  };
}

const menuItems = computed<SidebarItem[]>(() => {
  // Touch auth.profile so this computed re-runs when /auth/me resolves.
  const menus = auth.profile?.menus ?? [];
  return (
    menus
      // Dashboard is hardcoded above; skip the menu-tree duplicate.
      .filter((m) => m.path !== '/dashboard')
      .map(toItem)
      .filter((x): x is SidebarItem => x !== null)
  );
});

async function onCommand(cmd: string) {
  if (cmd === 'logout') {
    await auth.logout();
    router.push('/login');
  } else if (cmd === 'profile') {
    router.push('/profile');
  }
}
</script>

<style scoped>
.layout {
  height: 100vh;
}
.aside {
  background: #001529;
  transition: width 0.2s;
  overflow: hidden;
}
.logo {
  height: var(--app-header-height);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 600;
  letter-spacing: 1px;
}
.header {
  background: #fff;
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--app-border);
}
.main {
  background: var(--app-bg);
  padding: 16px;
}
.spacer {
  flex: 1;
}
.user {
  display: flex;
  align-items: center;
  cursor: pointer;
}
:deep(.el-menu) {
  border-right: 0;
}
</style>
