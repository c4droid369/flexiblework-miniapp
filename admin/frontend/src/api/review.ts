import type { PageData } from '@/utils/request';
import { del, get } from '@/utils/request';

// Mirror backend dto.ReviewResp. For the admin moderation page we expose
// every review across all orders; the backend currently returns an empty
// stream and admins drill in via /orders/:id/reviews.
export interface AdminReview {
  id: number;
  order_id: number;
  from_user_id: number;
  from_name: string;
  from_avatar: string;
  to_user_id: number;
  role: number;
  rating: number;
  content: string;
  tags: string[];
  created_at: string;
}

function qs(params: Record<string, unknown>): string {
  const usp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== '') usp.set(k, String(v));
  }
  return usp.toString();
}

export const reviewApi = {
  // Admin's all-reviews stream — backend placeholder returns []. We render
  // it as a guided message: admins use per-order reviews for content.
  list: (page = 1, size = 20) =>
    get<PageData<AdminReview>>(`/admin/reviews?${qs({ page, size })}`),
  remove: (id: number) => del(`/admin/reviews/${id}`),
};
