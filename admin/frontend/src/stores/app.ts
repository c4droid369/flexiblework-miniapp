import { defineStore } from 'pinia';
import { ref } from 'vue';

// Sidebar collapse + breadcrumb crumbs are global UI state that survives
// route changes. Pure UI concerns — no auth or domain logic here.
export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref<boolean>(false);
  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value;
  }
  return { sidebarCollapsed, toggleSidebar };
});
