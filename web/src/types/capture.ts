export interface StartCaptureRequest {
  interface?: string;
  bpf_filter?: string;
  max_packets?: number;
  duration_secs?: number;
  description?: string;
}

export interface CaptureSessionResponse {
  id: string;
  interface: string;
  bpf_filter: string;
  status: string;
  max_packets: number;
  duration_secs: number;
  description: string;
  packet_count: number;
  file_size: number;
  file_path: string;
  created_at: string;
  started_at?: string;
  stopped_at?: string;
  error_msg?: string;
}

export interface CaptureListResponse {
  sessions: CaptureSessionResponse[];
  total: number;
}
