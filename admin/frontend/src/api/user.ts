import type { PageData } from '@/utils/request';
import { del, download, get, post, put } from '@/utils/request';

export interface User {
  id: number;
  username: string;
  nickname: string;
  email: string;
  phone: string;
  avatar: string;
  status: 1 | 2;
  last_login_at?: string;
  last_login_ip: string;
  remark: string;
  created_at: string;
  updated_at: string;
  roles: Array<{ id: number; code: string; name: string }>;
}

export interface CreateUserReq {
  username: string;
  password: string;
  nickname?: string;
  email?: string;
  phone?: string;
  avatar?: string;
  remark?: string;
  role_ids?: number[];
}

export interface UpdateUserReq {
  nickname?: string;
  email?: string;
  phone?: string;
  avatar?: string;
  remark?: string;
  status?: 1 | 2;
  role_ids?: number[];
}

function qs(params: Record<string, unknown>): string {
  const usp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== '') usp.set(k, String(v));
  }
  return usp.toString();
}

export const userApi = {
  list: (page: number, size: number, keyword?: string) =>
    get<PageData<User>>(`/system/users?${qs({ page, size, keyword })}`),
  get: (id: number) => get<User>(`/system/users/${id}`),
  create: (body: CreateUserReq) => post<User>('/system/users', body),
  update: (id: number, body: UpdateUserReq) => put<User>(`/system/users/${id}`, body),
  remove: (id: number) => del(`/system/users/${id}`),
  batchRemove: (ids: number[]) => post<null>('/system/users/batch-delete', { ids }),
  resetPassword: (id: number, newPassword: string) =>
    post<null>(`/system/users/${id}/reset-password`, { new_password: newPassword }),
  changeStatus: (id: number, status: 1 | 2) => post<null>(`/system/users/${id}/status`, { status }),
  assignRoles: (id: number, roleIds: number[]) =>
    post<null>(`/system/users/${id}/roles`, { ids: roleIds }),
  exportExcel: (keyword?: string) => download(`/system/users/export/excel?${qs({ keyword })}`),
  exportCSV: (keyword?: string) => download(`/system/users/export/csv?${qs({ keyword })}`),
};
