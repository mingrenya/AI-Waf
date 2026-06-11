import { post, get } from './index'

export interface FTWResult {
  testId: string
  title: string
  passed: boolean
  statusCode: number
  wafHit: boolean
  wafRuleId: number
  wafMessage: string
  category: string
  error?: string
  durationMs: number
}

export interface FTWReport {
  id: string
  targetUrl: string
  totalTests: number
  passed: number
  failed: number
  falseNegs: number
  falsePoss: number
  blockRate: number
  results: FTWResult[]
  createdAt: string
  durationSec: number
}

const BASE = '/ftw'

export const ftwApi = {
  run: (targetUrl: string) => post<FTWReport>(`${BASE}/run`, { targetUrl }),
  files: () => get<string[]>(`${BASE}/files`),
  reports: (limit = 20) => get<FTWReport[]>(`${BASE}/reports?limit=${limit}`),
}
