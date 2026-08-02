import { post } from '@/utils/request';

export interface BroadcastReq {
  title: string;
  content: string;
  link?: string;
  type?: 1 | 2 | 3 | 4;
  // all|admin|student|employer — empty = all.
  user_type?: 'all' | 'admin' | 'student' | 'employer';
}

export const messageApi = {
  broadcast: (body: BroadcastReq) => post<number>('/admin/messages/broadcast', body),
};
