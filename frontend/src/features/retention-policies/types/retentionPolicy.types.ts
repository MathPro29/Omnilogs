export type RetentionMode = "DAILY" | "WEEKLY" | "MONTHLY" | "CUSTOM";
export type RetentionUnit =
  "DAY" | "WEEK" | "MONTH" | "YEAR" | "DAYS" | "WEEKS" | "MONTHS" | "YEARS";
export type FolderStructure =
  "7_DAILY_FOLDERS" | "4_WEEKLY_SUBFOLDERS" | "MONTHLY_FOLDERS" | "CUSTOM";
export type AutoPurgeAction = "DELETE" | "ARCHIVE_COLD" | "COMPRESS_GZIP";
export type StorageProvider = "R2" | "S3" | "GCS" | "MINIO" | "LOCAL";

export interface RetentionPolicy {
  policy_id: number;
  product_id: number;
  environment_id?: number;
  project_id?: number;
  name: string;
  description: string;
  retention_mode: RetentionMode;
  retention_unit: RetentionUnit;
  retention_value: number;
  total_days: number;
  folder_structure: FolderStructure;
  auto_purge_action: AutoPurgeAction;
  storage_provider: StorageProvider;
  bucket_name?: string;
  cron_schedule: string;
  schedule_timezone: string;
  is_active: boolean;
  last_purge_at?: string;
  next_purge_at?: string;
  created_at: string;
  updated_at: string;
  active_retention_value?: number;
  active_retention_unit?: RetentionUnit;
  archive_enabled?: boolean;
  archive_after_value?: number;
  archive_after_unit?: RetentionUnit;
  delete_active_after_archive?: boolean;
  archive_retention_value?: number;
  archive_retention_unit?: RetentionUnit;
  archive_never_delete?: boolean;
  apply_to_existing_logs?: boolean;
}

export interface CreateRetentionPolicyPayload {
  product_id: number;
  environment_id?: number;
  project_id?: number;
  name: string;
  description: string;
  retention_mode: RetentionMode;
  retention_unit: RetentionUnit;
  retention_value: number;
  folder_structure?: FolderStructure;
  auto_purge_action?: AutoPurgeAction;
  storage_provider?: StorageProvider;
  bucket_name?: string;
  cron_schedule?: string;
  schedule_timezone?: string;
  is_active?: boolean;
  active_retention_value?: number;
  active_retention_unit?: RetentionUnit;
  archive_enabled?: boolean;
  archive_after_value?: number;
  archive_after_unit?: RetentionUnit;
  delete_active_after_archive?: boolean;
  archive_retention_value?: number;
  archive_retention_unit?: RetentionUnit;
  archive_never_delete?: boolean;
  apply_to_existing_logs?: boolean;
}

export interface UpdateRetentionPolicyPayload {
  name?: string;
  description?: string;
  retention_mode?: RetentionMode;
  retention_unit?: RetentionUnit;
  retention_value?: number;
  folder_structure?: FolderStructure;
  auto_purge_action?: AutoPurgeAction;
  storage_provider?: StorageProvider;
  bucket_name?: string;
  cron_schedule?: string;
  schedule_timezone?: string;
  is_active?: boolean;
  environment_id?: number;
  active_retention_value?: number;
  active_retention_unit?: RetentionUnit;
  archive_enabled?: boolean;
  archive_after_value?: number;
  archive_after_unit?: RetentionUnit;
  delete_active_after_archive?: boolean;
  archive_retention_value?: number;
  archive_retention_unit?: RetentionUnit;
  archive_never_delete?: boolean;
  apply_to_existing_logs?: boolean;
}

export interface FolderNode {
  key: string;
  title: string;
  path: string;
  type: "root" | "month" | "week" | "day" | "file";
  folder_count: number;
  file_count: number;
  estimated_mb: number;
  status: "ACTIVE" | "ARCHIVED" | "TO_BE_PURGED";
  expiry_date?: string;
  children?: FolderNode[];
}

export interface SimulationResult {
  retention_mode: RetentionMode;
  total_retention_days: number;
  calculated_cutoff_date: string;
  total_folders_count: number;
  active_folders_count: number;
  purge_folders_count: number;
  estimated_savings_mb: number;
  tree_data: FolderNode;
}

export interface PolicyStats {
  total_policies: number;
  active_policies: number;
  total_storage_bytes: number;
  active_storage_bytes: number;
  archive_gzip_storage_bytes: number;
  ready_zip_storage_bytes: number;
  total_storage_gb: number;
  scheduled_purges_24h: number;
  total_folders_count: number;
}

export interface HistoricalBucket {
  period: string;
  date_from: string;
  date_to: string;
  oldest_log?: string;
  newest_log?: string;
  document_count: number;
  estimated_size_bytes: number;
  archive_status: string;
}

export interface HistoricalResponse {
  product_id: number;
  environment_id: number;
  oldest_log?: string;
  newest_log?: string;
  document_count: number;
  estimated_size_bytes: number;
  archive_status: string;
  buckets: HistoricalBucket[];
}

export interface RetentionPreviewResponse {
  historical: HistoricalResponse;
  archive_candidate_count: number;
  delete_candidate_count: number;
  delete_requires_verified_archive: boolean;
  active_cutoff?: string;
  archive_cutoff?: string;
}

export interface LogArchive {
  archive_id: string;
  product_id: number;
  environment_id?: number;
  policy_id?: number;
  backup_type: "POLICY" | "MANUAL";
  backup_tag: string;
  coverage_key?: string;
  storage_provider?: StorageProvider;
  bucket_name?: string;
  object_key?: string;
  category_id?: number;
  feature_id?: number;
  sub_feature_id?: number;
  date_from?: string;
  date_to?: string;
  document_count: number;
  original_size_bytes: number;
  compressed_size_bytes?: number;
  snapshot_repository?: string;
  snapshot_name?: string;
  snapshot_uuid?: string;
  snapshot_status?: string;
  snapshot_verified_at?: string;
  restore_index_pattern?: string;
  restore_status?: string;
  status?: string;
  exported_at?: string;
  verified_at?: string;
  deleted_from_active_at?: string;
}

export interface ArchiveJob {
  job_id: string;
  status: string;
  archive_ids?: string[];
  started_at?: string;
  completed_at?: string;
}

export interface ReadyArchiveDownload {
  download_id: string;
  product_id: number;
  environment_id: number;
  policy_id?: number;
  date_from?: string;
  date_to?: string;
  file_name: string;
  size_bytes: number;
  archive_count: number;
  status: "READY_TO_LOAD" | string;
  downloaded_at?: string;
  created_at?: string;
}

export interface CreateArchiveResponse {
  job: ArchiveJob;
  archives: LogArchive[];
}

export interface ImportArchiveResponse {
  imported_count: number;
  archives: LogArchive[];
}

export interface VerifyArchiveResponse {
  archive_id: string;
  status: string;
  checksum_valid: boolean;
  manifest_valid: boolean;
  snapshot_valid: boolean;
  document_count: number;
  deleted_active_logs: boolean;
}

export interface RestoreArchiveResponse {
  archive_id: string;
  mode: string;
  index_name: string;
  snapshot_restore?: boolean;
  restored_count: number;
  skipped_count: number;
  conflict_strategy?: string;
}
