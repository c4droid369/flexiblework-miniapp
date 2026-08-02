import { del, get, post, put } from '@/utils/request';
import type { MenuTree } from './auth';

export interface Menu extends Omit<MenuTree, 'children'> {}

export interface CreateMenuReq {
  parent_id: number;
  type: 1 | 2 | 3;
  name: string;
  title?: string;
  path?: string;
  component?: string;
  perm_code?: string;
  icon?: string;
  sort?: number;
  visible?: boolean;
}

export const menuApi = {
  tree: () => get<MenuTree[]>('/system/menus'),
  get: (id: number) => get<Menu>(`/system/menus/${id}`),
  create: (body: CreateMenuReq) => post<Menu>('/system/menus', body),
  update: (id: number, body: Partial<CreateMenuReq>) => put<Menu>(`/system/menus/${id}`, body),
  remove: (id: number) => del(`/system/menus/${id}`),
};
