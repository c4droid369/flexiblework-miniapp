import type { PageData } from '@/utils/request';
import { del, get, post, put } from '@/utils/request';

export interface Category {
  id: number;
  name: string;
  icon: string;
  sort: number;
  status: 1 | 2;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCategoryReq {
  name: string;
  icon?: string;
  sort?: number;
  status?: 1 | 2;
  description?: string;
}

export interface UpdateCategoryReq {
  name?: string;
  icon?: string;
  sort?: number;
  status?: 1 | 2;
  description?: string;
}

function qs(params: Record<string, unknown>): string {
  const usp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== '') usp.set(k, String(v));
  }
  return usp.toString();
}

export const categoryApi = {
  list: (status?: 1 | 2) => get<Category[]>(`/admin/categories?${qs({ status: status ?? '' })}`),
  create: (body: CreateCategoryReq) => post<Category>('/admin/categories', body),
  update: (id: number, body: UpdateCategoryReq) => put<Category>(`/admin/categories/${id}`, body),
  remove: (id: number) => del(`/admin/categories/${id}`),
};
