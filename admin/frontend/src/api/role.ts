import type { PageData } from '@/utils/request';
import { del, download, get, post, put } from '@/utils/request';

export interface Role {
  id: number;
  code: string;
  name: string;
  description: string;
  sort: number;
  status: 1 | 2;
  created_at: string;
  updated_at: string;
  menu_ids: number[];
}

export interface CreateRoleReq {
  code: string;
  name: string;
  description?: string;
  sort?: number;
  status?: 1 | 2;
  menu_ids?: number[];
}

function qs(params: Record<string, unknown>): string {
  const usp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== '') usp.set(k, String(v));
  }
  return usp.toString();
}

export const roleApi = {
  list: (page: number, size: number, keyword?: string) =>
    get<PageData<Role>>(`/system/roles?${qs({ page, size, keyword })}`),
  get: (id: number) => get<Role>(`/system/roles/${id}`),
  create: (body: CreateRoleReq) => post<Role>('/system/roles', body),
  update: (id: number, body: Partial<CreateRoleReq>) => put<Role>(`/system/roles/${id}`, body),
  remove: (id: number) => del(`/system/roles/${id}`),
  batchRemove: (ids: number[]) => post<null>('/system/roles/batch-delete', { ids }),
  assignMenus: (id: number, menuIds: number[]) =>
    post<null>(`/system/roles/${id}/menus`, { menu_ids: menuIds }),
  exportExcel: (keyword?: string) => download(`/system/roles/export/excel?${qs({ keyword })}`),
  exportCSV: (keyword?: string) => download(`/system/roles/export/csv?${qs({ keyword })}`),
};
