import { get, post } from '@/utils/request';

export interface LoginReq {
  username: string;
  password: string;
}
export interface LoginResp {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  permissions: string[];
}

export interface MenuTree {
  id: number;
  parent_id: number;
  type: 1 | 2 | 3;
  name: string;
  title: string;
  path: string;
  component: string;
  perm_code: string;
  icon: string;
  sort: number;
  visible: boolean;
  children?: MenuTree[];
}

export interface MeResp {
  id: number;
  username: string;
  nickname: string;
  email: string;
  phone: string;
  avatar: string;
  status: 1 | 2;
  last_login_at?: string;
  roles: Array<{ id: number; code: string; name: string }>;
  permissions: string[];
  menus: MenuTree[];
}

export const authApi = {
  login: (body: LoginReq) => post<LoginResp>('/auth/login', body),
  refresh: (refreshToken: string) =>
    post<LoginResp>('/auth/refresh', { refresh_token: refreshToken }),
  logout: (refreshToken?: string) =>
    post<null>('/auth/logout', refreshToken ? { refresh_token: refreshToken } : undefined),
  me: () => get<MeResp>('/auth/me'),
};
