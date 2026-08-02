import axios, {
  type AxiosError,
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from 'axios';
import { ElMessage } from 'element-plus';
import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
  setAccessToken,
  setRefreshToken,
} from '@/utils/auth';

// Envelope shape mirrored from backend internal/pkg/response/response.go.
export interface ApiEnvelope<T = unknown> {
  code: number;
  message: string;
  data: T;
}

export interface PageData<T> {
  list: T[];
  total: number;
  page: number;
  size: number;
}

// Single instance — every API module imports this.
const instance: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
});

// Attach bearer token on the way out.
instance.interceptors.request.use((cfg: InternalAxiosRequestConfig) => {
  const token = getAccessToken();
  if (token) cfg.headers.set('Authorization', `Bearer ${token}`);
  return cfg;
});

// Track refresh state so concurrent 401s don't fire N refresh calls.
let refreshing = false;
let pendingQueue: Array<(ok: boolean) => void> = [];

function notifyRefresh(ok: boolean) {
  pendingQueue.forEach((cb) => cb(ok));
  pendingQueue = [];
}

// Single response interceptor: unwrap envelope, normalize errors, refresh on 401.
instance.interceptors.response.use(
  (resp: AxiosResponse<ApiEnvelope>) => {
    const env = resp.data;
    if (env && typeof env === 'object' && 'code' in env) {
      if (env.code === 0) return resp;
      const msg = env.message || 'request failed';
      ElMessage.error(msg);
      return Promise.reject(new Error(msg));
    }
    return resp;
  },
  async (err: AxiosError<ApiEnvelope>) => {
    const status = err.response?.status;
    const original = err.config as InternalAxiosRequestConfig & { _retry?: boolean };

    if (status === 401 && !original._retry) {
      const refreshToken = getRefreshToken();
      if (!refreshToken) {
        clearTokens();
        window.location.href = '/login';
        return Promise.reject(err);
      }
      if (refreshing) {
        return new Promise((resolve, reject) => {
          pendingQueue.push((ok) => (ok ? resolve(instance(original)) : reject(err)));
        });
      }
      refreshing = true;
      original._retry = true;
      try {
        const r = await axios.post<ApiEnvelope<{ access_token: string; refresh_token: string }>>(
          `${import.meta.env.VITE_API_BASE_URL || '/api/v1'}/auth/refresh`,
          { refresh_token: refreshToken },
        );
        const data = r.data.data;
        setAccessToken(data.access_token);
        setRefreshToken(data.refresh_token);
        notifyRefresh(true);
        original.headers.set('Authorization', `Bearer ${data.access_token}`);
        return instance(original);
      } catch {
        clearTokens();
        notifyRefresh(false);
        window.location.href = '/login';
        return Promise.reject(err);
      } finally {
        refreshing = false;
      }
    }

    const msg = err.response?.data?.message || err.message || 'network error';
    ElMessage.error(msg);
    return Promise.reject(err);
  },
);

// Strongly typed helper. Pass the expected inner data type as T.
export async function get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
  const r = await instance.get<unknown, AxiosResponse<ApiEnvelope<T>>>(url, { params });
  return r.data.data;
}

export async function post<T>(url: string, body?: unknown): Promise<T> {
  const r = await instance.post<unknown, AxiosResponse<ApiEnvelope<T>>>(url, body);
  return r.data.data;
}

export async function put<T>(url: string, body?: unknown): Promise<T> {
  const r = await instance.put<unknown, AxiosResponse<ApiEnvelope<T>>>(url, body);
  return r.data.data;
}

export async function del<T = null>(url: string): Promise<T> {
  const r = await instance.delete<unknown, AxiosResponse<ApiEnvelope<T>>>(url);
  return r.data.data;
}

// Direct download (Excel/CSV) — uses raw axios so the response stream isn't
// JSON-decoded. Caller is responsible for triggering save.
export async function download(url: string, params?: Record<string, unknown>): Promise<Blob> {
  const r = await instance.get(url, { params, responseType: 'blob' });
  return r.data;
}

export default instance;
