export interface Lead {
  id: string;
  received_at: string;
  source: string;
  sender: string;
  recipient: string;
  subject: string;
  raw_payload?: string;
  parsed_body: string;
  extracted_entities: string;
  company_name: string;
  domain: string;
  contact_name: string;
  contact_email: string;
  contact_phone: string;
  requested_seats: number;
  budget: number;
  urgency: string;
  icp_score: number;
  qualification_tier: 'Enterprise Hot Lead' | 'Mid-Market Qualified' | 'SMB / Self-Serve' | 'Unqualified / Low Intent';
  dedup_status: 'new_account' | 'existing_matched' | 'manual_review';
  matched_account_id?: string;
  matched_account_name?: string;
  matched_account_owner?: string;
  outreach_draft: string;
  status: 'received' | 'parsed' | 'extracted' | 'scored' | 'synced' | 'failed';
  latency_ms: number;
  created_at: string;
  updated_at: string;
}

export interface ExtractedEntitiesData {
  company_name: string;
  domain: string;
  contact_name: string;
  job_title: string;
  seniority: string;
  contact_email: string;
  contact_phone: string;
  requested_products: string[];
  requested_seats: number;
  budget: number;
  currency: string;
  urgency: string;
  tech_stack: string[];
  pain_points: string;
  confidence_score: number;
  extraction_method: string;
}

export interface SyncLog {
  id: string;
  lead_id: string;
  crm_type: string;
  action: string;
  status: string;
  status_code: number;
  request_payload: string;
  response_payload: string;
  duration_ms: number;
  error_message?: string;
  created_at: string;
}

export interface AuditEvent {
  id: string;
  lead_id: string;
  stage: string;
  event: string;
  details: string;
  created_at: string;
}

export interface AccountRecord {
  id: string;
  domain: string;
  company_name: string;
  owner_name: string;
  owner_email: string;
  crm_id: string;
  crm_type: string;
  created_at: string;
  updated_at: string;
}

export interface StatsSummary {
  total_leads: number;
  enterprise_hot_leads: number;
  mid_market_leads: number;
  smb_leads: number;
  unqualified_leads: number;
  deduplicated_accounts: number;
  successful_syncs: number;
  average_latency_ms: number;
}

export interface HealthInfo {
  status: string;
  uptime_seconds: number;
  memory_alloc_mb: string;
  memory_sys_mb: string;
  memory_target: string;
  within_budget: boolean;
  goroutines: number;
  crm_mode: string;
  extraction_mode: string;
}
