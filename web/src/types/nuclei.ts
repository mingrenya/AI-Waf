export interface ScanRequest {
  site_id: string;
  target_url: string;
  templates: string[];
  severity: string;
}

export interface ScanTaskResponse {
  id: string;
  site_id: string;
  target_url: string;
  status: string;
  findings: number;
  total: number;
  created_at: string;
}

export interface TemplateInfo {
  path: string;
  name: string;
}
