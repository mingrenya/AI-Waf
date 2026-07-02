import api from './index';
import type { StartCaptureRequest, CaptureSessionResponse, CaptureListResponse } from '@/types/capture';

const BASE = '/capture';

export const startCapture = (req: StartCaptureRequest) =>
  api.post<{ data: CaptureSessionResponse }>(`${BASE}/start`, req);

export const stopCapture = (id: string) =>
  api.post<{ data: CaptureSessionResponse }>(`${BASE}/${id}/stop`);

export const getSession = (id: string) =>
  api.get<{ data: CaptureSessionResponse }>(`${BASE}/${id}`);

export const listSessions = () =>
  api.get<{ data: CaptureListResponse }>(`${BASE}/sessions`);

export const getDownloadUrl = (id: string) =>
  `${api.defaults.baseURL}${BASE}/${id}/download`;
