import api from './index';

const BASE = '/nuclei';

export const startScan = (req: { site_id: string; target_url: string; templates?: string[]; severity?: string }) =>
  api.post(`${BASE}/scan`, req);

export const getTask = (id: string) =>
  api.get(`${BASE}/scan/${id}`);

export const cancelTask = (id: string) =>
  api.post(`${BASE}/scan/${id}/cancel`);

export const listTasks = () =>
  api.get(`${BASE}/tasks`);

export const listTemplates = () =>
  api.get(`${BASE}/templates`);
