import type { PageData } from '@/utils/request';
import { get } from '@/utils/request';

// Mirror backend dto.OrderResp.
export interface AdminOrder {
  id: number;
  order_no: string;
  job_id: number;
  job_title: string;
  application_id: number;
  employer_id: number;
  employer_name: string;
  student_id: number;
  student_name: string;
  amount: number;
  status: number;
  pay_method: string;
  paid_at: string | null;
  started_at: string | null;
  completed_at: string | null;
  confirmed_at: string | null;
  settled_at: string | null;
  work_proof: string[];
  cancel_reason: string;
  created_at: string;
}

const ORDER_STATUS_TEXT: Record<number, string> = {
  1: '待支付', 2: '已支付', 3: '进行中', 4: '待确认完成',
  5: '已结算', 6: '已取消', 7: '已退款',
};

const ORDER_STATUS_TAG: Record<number, '' | 'success' | 'warning' | 'info' | 'primary' | 'danger'> = {
  1: 'warning', 2: 'info', 3: 'primary', 4: 'warning',
  5: 'success', 6: 'info', 7: 'info',
};

function qs(params: Record<string, unknown>): string {
  const usp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== '') usp.set(k, String(v));
  }
  return usp.toString();
}

export const orderApi = {
  list: (page = 1, size = 20, status?: number) =>
    get<PageData<AdminOrder>>(`/admin/orders?${qs({ page, size, status: status ?? '' })}`),
};

export { ORDER_STATUS_TEXT, ORDER_STATUS_TAG };
