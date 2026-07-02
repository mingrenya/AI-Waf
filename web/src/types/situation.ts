export type AttackStage =
  | 'unknown'
  | 'reconnaissance'
  | 'scanning'
  | 'exploitation'
  | 'lateral_movement'
  | 'command_and_control'
  | 'exfiltration';

export interface SituationOverview {
  active_chains: number;
  total_chains_24h: number;
  total_attackers_24h: number;
  overall_risk_score: number;
  risk_trend: 'rising' | 'stable' | 'falling';
  top_attack_types: CountItem[];
  top_attacker_ips: CountItem[];
  top_target_sites: CountItem[];
  by_country: CountItem[];
}

export interface CountItem {
  label: string;
  count: number;
}

export interface ChainSummary {
  id: string;
  source_ip: string;
  geo_country: string;
  stages: string[];
  risk_score: number;
  first_seen: string;
  last_seen: string;
  active: boolean;
}

export interface ChainListResponse {
  chains: ChainSummary[];
  total: number;
  page: number;
  page_size: number;
}

export interface ChainStageItem {
  stage: string;
  technique: string;
  detected_at: string;
  confidence: number;
  evidence: string[];
}

export interface AttackerProfileDetail {
  source_ip: string;
  geo_country: string;
  geo_city?: string;
  total_attacks: number;
  unique_attack_types: number;
  top_attack_type: string;
  unique_target_sites: number;
  active_hours: number[];
  attack_phase: string;
  tools_identified: string;
  is_automated: boolean;
  is_persistent: boolean;
  risk_score: number;
  risk_label: string;
  first_seen: string;
  last_seen: string;
  recent_events?: LogEventItem[];
}

export interface LogEventItem {
  id: string;
  attack_type: string;
  severity: string;
  action: string;
  site_domain: string;
  correlation_id: string;
  timestamp: string;
}

export interface ChainDetail {
  id: string;
  source_ip: string;
  geo_country: string;
  stages: ChainStageItem[];
  correlation_ids: string[];
  risk_score: number;
  risk_label: string;
  first_seen: string;
  last_seen: string;
  active: boolean;
  attacker_profile?: AttackerProfileDetail;
}

export interface AttackerSummary {
  source_ip: string;
  geo_country: string;
  total_attacks: number;
  unique_attack_types: number;
  top_attack_type: string;
  attack_phase: string;
  risk_score: number;
  risk_label: string;
  last_seen: string;
}

export interface TrendPoint {
  timestamp: number;
  total_events: number;
  blocked_count: number;
  detect_count: number;
  unique_ips: number;
}

export interface TrendResponse {
  timeline: TrendPoint[];
  frequent_types: CountItem[];
  active_attackers: number;
  new_chains_24h: number;
}

export interface SituationRule {
  id: string;
  name: string;
  stage: string;
  logql: string;
  interval_seconds: number;
  threshold: number;
  severity: string;
  mitre_tactic: string;
  mitre_technique: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface QuickActionRequest {
  source_ip: string;
  action: 'block' | 'blacklist' | 'both';
  duration_hours: number;
  reason: string;
  correlation_id?: string;
}

export interface QuickActionResponse {
  success: boolean;
  source_ip: string;
  action: string;
  blocked: boolean;
  blacklisted: boolean;
  note: string;
}

export interface WSSituationMessage {
  type: 'situation:alert' | 'situation:update' | 'situation:attack' | 'situation:quick_action';
  payload: unknown;
  time: string;
}
