import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@/api/auth', () => ({
  authApi: {
    login: vi.fn(),
    refresh: vi.fn(),
    logout: vi.fn(),
    me: vi.fn(),
  },
}));

import { createPinia, setActivePinia } from 'pinia';
import { authApi } from '@/api/auth';
import { useAuthStore } from '@/stores/auth';
import { clearTokens, getAccessToken, getRefreshToken } from '@/utils/auth';

const mockedApi = vi.mocked(authApi, true);

describe('stores/auth', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    clearTokens();
    vi.clearAllMocks();
  });

  describe('login', () => {
    it('persists tokens and hydrates profile', async () => {
      mockedApi.login.mockResolvedValueOnce({
        access_token: 'a-1',
        refresh_token: 'r-1',
        expires_in: 3600,
        token_type: 'Bearer',
        permissions: ['user:view', 'user:create'],
      });
      mockedApi.me.mockResolvedValueOnce({
        id: 1,
        username: 'admin',
        nickname: 'Admin',
        email: '',
        phone: '',
        avatar: '',
        status: 1,
        roles: [{ id: 1, code: 'super_admin', name: 'Super' }],
        permissions: ['user:view', 'user:create'],
        menus: [],
      });

      const store = useAuthStore();
      await store.login('admin', 'admin123');

      expect(mockedApi.login).toHaveBeenCalledWith({ username: 'admin', password: 'admin123' });
      expect(mockedApi.me).toHaveBeenCalled();
      expect(store.token).toBe('a-1');
      expect(store.refreshToken).toBe('r-1');
      expect(store.profile?.username).toBe('admin');
      expect(getAccessToken()).toBe('a-1');
      expect(getRefreshToken()).toBe('r-1');
      expect(store.isAuthenticated).toBe(true);
    });
  });

  describe('hasPerm', () => {
    it('super_admin bypasses perm check', async () => {
      const store = useAuthStore();
      mockedApi.me.mockResolvedValueOnce({
        id: 1,
        username: 'x',
        nickname: '',
        email: '',
        phone: '',
        avatar: '',
        status: 1,
        roles: [{ id: 1, code: 'super_admin', name: 'Super' }],
        permissions: [],
        menus: [],
      });
      await store.fetchMe();
      expect(store.hasPerm('anything:at-all')).toBe(true);
    });

    it('non-super checks membership', async () => {
      const store = useAuthStore();
      mockedApi.me.mockResolvedValueOnce({
        id: 2,
        username: 'u',
        nickname: '',
        email: '',
        phone: '',
        avatar: '',
        status: 1,
        roles: [{ id: 2, code: 'common', name: 'User' }],
        permissions: ['user:view'],
        menus: [],
      });
      await store.fetchMe();
      expect(store.hasPerm('user:view')).toBe(true);
      expect(store.hasPerm('user:delete')).toBe(false);
    });

    it('empty perm string is always allowed', async () => {
      const store = useAuthStore();
      // No profile loaded yet, but empty perm should still return true.
      expect(store.hasPerm('')).toBe(true);
    });
  });

  describe('logout', () => {
    it('clears local state even when backend logout fails', async () => {
      mockedApi.login.mockResolvedValueOnce({
        access_token: 'a',
        refresh_token: 'r',
        expires_in: 60,
        token_type: 'Bearer',
        permissions: [],
      });
      mockedApi.me.mockResolvedValueOnce({
        id: 1,
        username: 'a',
        nickname: '',
        email: '',
        phone: '',
        avatar: '',
        status: 1,
        roles: [],
        permissions: [],
        menus: [],
      });
      const store = useAuthStore();
      await store.login('a', 'p');

      mockedApi.logout.mockRejectedValueOnce(new Error('boom'));
      await store.logout();

      expect(store.token).toBe('');
      expect(store.profile).toBeNull();
      expect(getAccessToken()).toBe('');
      expect(getRefreshToken()).toBe('');
    });
  });
});
