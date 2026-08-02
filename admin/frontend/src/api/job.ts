import type { PageData } from '@/utils/request';
import { get, post } from '@/utils/request';

// Job list row — mirrors backend dto.JobResp.
export interface AdminJob {
  id: number;
  employer_id: number;
  employer_name: string;
  category_id: number;
  category_name: string;
  title: string;
  cover: string;
  description: string;
  requirements: string;
  salary_type: number;
  salary_min: number;
  salary_max: number;
  salary_unit: string;
  salary_text: string;
  location: string;
  work_date_type: number;
  work_date_start: string | null;
  work_date_end: string | null;
  work_time_start: string;
  work_time_end: string;
  recruit_count: number;
  gender_requirement: number;
  settlement_type: number;
  status: number;
  audit_remark: string;
  audited_at: string | null;
  view_count: number;
  apply_count: number;
  created_at: string;
  updated_at: string;
}

export interface AuditJobReq {
  action: 2 | 4; // 2=通过 4=拒绝
  remark?: string;
}

function qs(params: Record<string, unknown>): string {
  const usp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== '') usp.set(k, String(v));
  }
  return usp.toString();
}

export const jobApi = {
  // Admin: list pending jobs. status param is required by the backend to
  // scope to 待审核 (1).
  pending: (page = 1, size = 20) =>
    get<PageData<AdminJob>>(`/admin/jobs?${qs({ page, size, status: 1 })}`),
  audit: (id: number, body: AuditJobReq) => post<AdminJob>(`/admin/jobs/${id}/audit`, body),
};
