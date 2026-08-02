import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { authApi, type MeResp } from '@/api/auth';
import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
  setAccessToken,
  setRefreshToken,
} from '@/utils/auth';

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(getAccessToken());
  const refreshToken = ref<string>(getRefreshToken());
  const profile = ref<MeResp | null>(null);

  const isAuthenticated = computed(() => !!token.value);
  const permissions = computed<Set<string>>(() => new Set(profile.value?.permissions ?? []));

  async function login(username: string, password: string) {
    const r = await authApi.login({ username, password });
    token.value = r.access_token;
    refreshToken.value = r.refresh_token;
    setAccessToken(r.access_token);
    setRefreshToken(r.refresh_token);
    await fetchMe();
  }

  async function fetchMe() {
    profile.value = await authApi.me();
  }

  async function logout() {
    try {
      await authApi.logout(refreshToken.value || undefined);
    } catch {
      // even if backend logout fails, clear local state
    }
    token.value = '';
    refreshToken.value = '';
    profile.value = null;
    clearTokens();
  }

  function hasPerm(code: string): boolean {
    if (!code) return true;
    if (profile.value?.roles.some((r) => r.code === 'super_admin')) return true;
    return permissions.value.has(code);
  }

  return {
    token,
    refreshToken,
    profile,
    isAuthenticated,
    permissions,
    login,
    fetchMe,
    logout,
    hasPerm,
  };
});
