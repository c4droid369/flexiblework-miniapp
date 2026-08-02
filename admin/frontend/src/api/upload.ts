import type { PageData } from '@/utils/request';
import request, { del, get } from '@/utils/request';

export interface FileInfo {
  id: number;
  name: string;
  original_name: string;
  path: string;
  size: number;
  mime_type: string;
  storage: 'local' | string;
  uploader_id: number;
  created_at: string;
}

export interface UploadResp {
  id: number;
  url: string;
  name: string;
  size: number;
}

// request() is the raw axios instance; we use it directly for multipart
// because FormData lets axios compute the Content-Type (with boundary).
async function upload(file: File): Promise<UploadResp> {
  const fd = new FormData();
  fd.append('file', file);
  const r = await request.post<
    unknown,
    { data: { code: number; message: string; data: UploadResp } }
  >('/upload', fd);
  return r.data.data;
}

function buildQuery(params: Record<string, unknown>): string {
  const usp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === '') continue;
    usp.set(k, String(v));
  }
  return usp.toString();
}

export const uploadApi = {
  upload,
  list: (page: number, size: number, keyword?: string) =>
    get<PageData<FileInfo>>(`/files-list?${buildQuery({ page, size, keyword })}`),
  remove: (id: number) => del(`/files-list/${id}`),
};
