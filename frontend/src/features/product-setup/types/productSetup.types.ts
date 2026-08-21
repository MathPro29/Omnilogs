export type StepStatus = "not_started" | "in_progress" | "completed" | "needs_attention";

export type SetupStepKey = 
  | "product"
  | "projects"
  | "features"
  | "review"
  | "api_key"
  | "connect_logs"
  | "complete";

export interface SetupStepInfo {
  key: SetupStepKey;
  stepNumber: number;
  title: string;
  subtitle: string;
  status: StepStatus;
}

export interface EnvironmentInfo {
  environment_id?: number;
  environment_code: string;
  environment_name: string;
  environment_type: "DEVELOPMENT" | "PRODUCTION" | "TEST" | "QA" | "STAGING" | "UAT";
  is_default: boolean;
  is_active: boolean;
}

export interface ProjectInfo {
  project_id?: number;
  id?: string; // local UI id
  project_name: string;
  project_code?: string;
  description?: string;
  is_active: boolean;
  display_order: number;
}

export interface FeatureNode {
  category_id?: number;
  id: string; // local UI id
  project_id: number | string;
  parent_id: number | string | null;
  category_name: string;
  category_code: string;
  description?: string;
  is_active: boolean;
  display_order: number;
  children?: FeatureNode[];
}

export interface ProductSetupDraft {
  product_id?: number;
  product_name: string;
  product_code: string;
  description: string;
  owner_id?: number;
  owner_name?: string;
  is_active: boolean;
  setup_status: "DRAFT" | "STRUCTURE_INCOMPLETE" | "READY_FOR_API_KEY" | "WAITING_FOR_FIRST_LOG" | "ACTIVE";
  environments: EnvironmentInfo[];
  projects: ProjectInfo[];
  features: FeatureNode[];
}

export interface SetupStatusData {
  product_id: number;
  product_name: string;
  product_code: string;
  description?: string;
  setup_status: string;
  current_step: number;
  progress_percent: number;
  environments: EnvironmentInfo[];
  projects_count: number;
  features_count: number;
  sub_features_count: number;
  total_nodes_count: number;
  api_keys_count: number;
  test_log_received: boolean;
  validation_errors?: string[];
}

export interface ExistingApiKeyInfo {
  key_id: number;
  key_name: string;
  key_prefix: string;
  environment_id?: number;
  is_active: boolean;
  created_at?: string;
}
export interface GeneratedApiKeyInfo {
  key_id: number;
  key_name: string;
  raw_key: string;
  key_prefix: string;
  environment_code: string;
  product_code: string;
  expires_at?: string;
  permissions: string[];
}

export interface TestLogPayload {
  timestamp: string;
  level: "INFO" | "WARN" | "ERROR" | "DEBUG";
  message: string;
  project_id: number;
  category_id: number;
  project_code?: string;
  feature_code?: string;
  sub_feature_code?: string;
  custom_fields?: Record<string, unknown>;
  service: string;
  metadata: Record<string, unknown>;
}

export interface TestLogResult {
  log_id?: string;
  timestamp?: string;
  status: "waiting" | "sending" | "received" | "failed";
  error_message?: string;
  details?: Record<string, unknown>;
}


