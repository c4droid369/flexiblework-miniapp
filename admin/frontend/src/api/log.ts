import type { PageData } from '@/utils/request';
import { get, post } from '@/utils/request';

export interface OperationLog {
  id: number;
  user_id: number;
  username: string;
  action: string;
  method: string;
  path: string;
  ip: string;
  user_agent: string;
  request_body: string;
  response_status: number;
  latency_ms: number;
  error_message: string;
  created_at: string;
}

function qs(params: Record<string, unknown>): string {
  const usp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== '') usp.set(k, String(v));
  }
  return usp.toString();
}

export const logApi = {
  list: (page: number, size: number, keyword?: string, action?: string) =>
    get<PageData<OperationLog>>(`/system/logs?${qs({ page, size, keyword, action })}`),
  batchRemove: (ids: number[]) => post<null>('/system/logs/batch-delete', { ids }),
};
