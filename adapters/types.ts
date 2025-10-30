/**
 * Deploy Tunnel Adapter Bridge Protocol Types
 * Based on bridge_spec.json v1.0.0
 */

export type Provider = 'vercel' | 'cloudflare' | 'render' | 'netlify';
export type AuthType = 'oauth' | 'token' | 'api_key';
export type DeploymentStatus = 'queued' | 'building' | 'ready' | 'error';
export type RecordType = 'A' | 'AAAA' | 'CNAME' | 'TXT';
export type EnvTarget = 'production' | 'preview' | 'development';

export type ErrorCode =
  | 'AUTH_FAILED'
  | 'AUTH_REQUIRED'
  | 'PROVIDER_ERROR'
  | 'NETWORK_ERROR'
  | 'INVALID_PARAMS'
  | 'NOT_FOUND'
  | 'RATE_LIMITED'
  | 'UNSUPPORTED'
  | 'TIMEOUT'
  | 'UNKNOWN';

export interface BridgeError {
  code: ErrorCode;
  message: string;
  recoverable: boolean;
  details?: Record<string, unknown>;
}

export interface BridgeResponse<T = unknown> {
  ok: boolean;
  data?: T;
  error?: BridgeError;
  adapter_version: string;
}

// Command: auth:start
export interface AuthStartParams {
  provider: Provider;
  callback_url?: string;
}

export interface AuthStartData {
  auth_url?: string;
  token?: string;
  expires_at?: number;
}

// Command: auth:refresh
export interface AuthRefreshParams {
  provider: Provider;
  refresh_token: string;
}

export interface AuthRefreshData {
  token: string;
  expires_at: number;
}

// Command: fetch:config
export interface FetchConfigParams {
  provider: Provider;
  token: string;
  project_id?: string;
}

export interface EnvVar {
  key: string;
  value: string;
  target: EnvTarget[];
}

export interface FetchConfigData {
  project: {
    id: string;
    name: string;
    domain: string;
    framework?: string;
  };
  build: {
    command: string;
    output_dir: string;
    install_command?: string;
  };
  env: EnvVar[];
}

// Command: sync:env
export interface SyncEnvParams {
  provider: Provider;
  token: string;
  project_id: string;
  env_vars: EnvVar[];
}

export interface SyncEnvData {
  synced: number;
  failed: string[];
}

// Command: deploy:preview
export interface DeployPreviewParams {
  provider: Provider;
  token: string;
  project_id: string;
  branch?: string;
  env?: Record<string, string>;
}

export interface DeployPreviewData {
  deployment_id: string;
  url: string;
  status: DeploymentStatus;
  build_time?: number;
}

// Command: dns:update
export interface DnsUpdateParams {
  provider: Provider;
  token: string;
  domain: string;
  record_type: RecordType;
  record_name: string;
  record_value: string;
  ttl?: number;
}

export interface DnsUpdateData {
  record_id: string;
  previous_value?: string;
  propagation_time: number;
}

// Command: dns:rollback
export interface DnsRollbackParams {
  provider: Provider;
  token: string;
  record_id: string;
  rollback_to: string;
}

export interface DnsRollbackData {
  restored: boolean;
  current_value: string;
}

// Command: capabilities
export interface CapabilitiesData {
  adapter_name: string;
  adapter_version: string;
  supported_verbs: string[];
  auth_type: AuthType;
  features: {
    dns_management: boolean;
    preview_deployments: boolean;
    env_variables: boolean;
    build_logs: boolean;
  };
}

// Base adapter interface
export interface Adapter {
  capabilities(): Promise<BridgeResponse<CapabilitiesData>>;
  authStart(params: AuthStartParams): Promise<BridgeResponse<AuthStartData>>;
  authRefresh(params: AuthRefreshParams): Promise<BridgeResponse<AuthRefreshData>>;
  fetchConfig(params: FetchConfigParams): Promise<BridgeResponse<FetchConfigData>>;
  syncEnv(params: SyncEnvParams): Promise<BridgeResponse<SyncEnvData>>;
  deployPreview(params: DeployPreviewParams): Promise<BridgeResponse<DeployPreviewData>>;
  dnsUpdate(params: DnsUpdateParams): Promise<BridgeResponse<DnsUpdateData>>;
  dnsRollback(params: DnsRollbackParams): Promise<BridgeResponse<DnsRollbackData>>;
}
