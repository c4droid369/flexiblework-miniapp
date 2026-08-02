import { get, post } from '@/utils/request';

// Mirror backend dto.EmployerCertListItem / dto.StudentCertListItem.
export interface EmployerCertItem {
  user_id: number;
  username: string;
  nickname: string;
  company_name: string;
  contact_name: string;
  contact_phone: string;
  business_license_no: string;
  business_license_img: string;
  cert_status: number; // 1=审核中 2=已通过 3=已拒绝
  cert_remark: string;
  created_at: string;
}

export interface StudentCertItem {
  user_id: number;
  username: string;
  nickname: string;
  real_name: string;
  school: string;
  college: string;
  major: string;
  student_no: string;
  id_card_front: string;
  id_card_back: string;
  student_card: string;
  cert_status: number;
  cert_remark: string;
  created_at: string;
}

export interface CertAuditReq {
  action: 2 | 3; // 2=通过 3=拒绝
  remark?: string;
}

export const certApi = {
  listPendingStudent: () => get<StudentCertItem[]>('/admin/student-certifications'),
  auditStudent: (userId: number, body: CertAuditReq) =>
    post<null>(`/admin/student-certifications/${userId}/audit`, body),
  listPendingEmployer: () => get<EmployerCertItem[]>('/admin/employer-certifications'),
  auditEmployer: (userId: number, body: CertAuditReq) =>
    post<null>(`/admin/employer-certifications/${userId}/audit`, body),
  // Campus agent — reuses EmployerCertListItem shape (admin UI doesn't
  // distinguish employer vs agent at a glance).
  listPendingAgent: () => get<EmployerCertItem[]>('/admin/agent-certifications'),
  auditAgent: (userId: number, body: CertAuditReq) =>
    post<null>(`/admin/agent-certifications/${userId}/audit`, body),
};
