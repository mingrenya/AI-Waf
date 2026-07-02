import api from './index';
import type {
  SituationOverview,
  ChainListResponse,
  ChainDetail,
  AttackerSummary,
  AttackerProfileDetail,
  TrendResponse,
  SituationRule,
  QuickActionRequest,
  QuickActionResponse,
} from '@/types/situation';

const BASE = '/situation';

export const getOverview = () =>
  api.get<{ data: SituationOverview }>(`${BASE}/overview`);

export const listChains = (params: {
  source_ip?: string;
  stage?: string;
  active?: boolean;
  page?: number;
  page_size?: number;
}) => api.get<{ data: ChainListResponse }>(`${BASE}/chains`, { params });

export const getChainDetail = (id: string) =>
  api.get<{ data: ChainDetail }>(`${BASE}/chains/${id}`);

export const listAttackers = (params: {
  page?: number;
  page_size?: number;
  sort_by?: string;
  risk_label?: string;
}) => api.get<{ data: { attackers: AttackerSummary[]; total: number } }>(
  `${BASE}/attackers`,
  { params },
);

export const getAttackerProfile = (ip: string) =>
  api.get<{ data: AttackerProfileDetail }>(`${BASE}/attackers/${encodeURIComponent(ip)}`);

export const getTrends = (duration = '24h') =>
  api.get<{ data: TrendResponse }>(`${BASE}/trends`, { params: { duration } });

export const listRules = () =>
  api.get<{ data: SituationRule[] }>(`${BASE}/rules`);

export const createRule = (rule: Omit<SituationRule, 'id' | 'created_at' | 'updated_at'>) =>
  api.post(`${BASE}/rules`, rule);

export const updateRule = (id: string, rule: Partial<SituationRule>) =>
  api.put(`${BASE}/rules/${id}`, rule);

export const deleteRule = (id: string) =>
  api.delete(`${BASE}/rules/${id}`);

export const quickAction = (req: QuickActionRequest) =>
  api.post<{ data: QuickActionResponse }>(`${BASE}/quick-action`, req);
