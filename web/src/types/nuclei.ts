export interface ScanRequest {
  site_id: string;
  target_url: string;
  templates: string[];
  severity: string;
}

export interface NucleiFinding {
  template_id: string;
  name: string;
  severity: string;
  matched_at: string;
  curl_command: string;
  extracted_results: string[];
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

export interface ScanTaskDetail {
  id: string;
  site_id: string;
  target_url: string;
  status: string;
  findings: NucleiFinding[];
  total: number;
  created_at: string;
  started_at?: string;
  completed_at?: string;
}

export interface TemplateInfo {
  path: string;
  name: string;
}
