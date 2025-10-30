#!/usr/bin/env bun

import { BaseAdapter } from '../base';
import type {
  BridgeResponse,
  CapabilitiesData,
  AuthStartParams,
  AuthStartData,
  AuthRefreshParams,
  AuthRefreshData,
  FetchConfigParams,
  FetchConfigData,
  SyncEnvParams,
  SyncEnvData,
  DeployPreviewParams,
  DeployPreviewData,
  DnsUpdateParams,
  DnsUpdateData,
  DnsRollbackParams,
  DnsRollbackData,
} from '../types';

const VERCEL_API_BASE = 'https://api.vercel.com';

class VercelAdapter extends BaseAdapter {
  async capabilities(): Promise<BridgeResponse<CapabilitiesData>> {
    return this.success({
      adapter_name: 'vercel',
      adapter_version: this.version,
      supported_verbs: [
        'capabilities',
        'auth:start',
        'fetch:config',
        'sync:env',
        'deploy:preview',
        'dns:update',
        'dns:rollback',
      ],
      auth_type: 'token',
      features: {
        dns_management: true,
        preview_deployments: true,
        env_variables: true,
        build_logs: true,
      },
    });
  }

  async authStart(params: AuthStartParams): Promise<BridgeResponse<AuthStartData>> {
    // Vercel uses personal access tokens
    // Guide user to create one at https://vercel.com/account/tokens
    return this.success({
      auth_url: 'https://vercel.com/account/tokens',
      token: undefined,
      expires_at: undefined,
    });
  }

  async authRefresh(params: AuthRefreshParams): Promise<BridgeResponse<AuthRefreshData>> {
    // Vercel tokens don't expire, so refresh is not needed
    return this.unsupported('auth:refresh');
  }

  async fetchConfig(params: FetchConfigParams): Promise<BridgeResponse<FetchConfigData>> {
    const { token, project_id } = params;

    try {
      // If no project_id, list projects and let user choose
      if (!project_id) {
        const response = await fetch(`${VERCEL_API_BASE}/v9/projects`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          const error = await response.json();
          return this.error({
            code: response.status === 401 ? 'AUTH_FAILED' : 'PROVIDER_ERROR',
            message: error.error?.message || 'Failed to fetch projects',
            recoverable: response.status === 401,
            details: error,
          });
        }

        const data = await response.json();
        return this.error({
          code: 'INVALID_PARAMS',
          message: 'Multiple projects found. Please specify project_id.',
          recoverable: true,
          details: {
            projects: data.projects.map((p: any) => ({
              id: p.id,
              name: p.name,
            })),
          },
        });
      }

      // Fetch specific project
      const projectResponse = await fetch(`${VERCEL_API_BASE}/v9/projects/${project_id}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!projectResponse.ok) {
        const error = await projectResponse.json();
        return this.error({
          code: projectResponse.status === 404 ? 'NOT_FOUND' : 'PROVIDER_ERROR',
          message: error.error?.message || 'Failed to fetch project',
          recoverable: false,
          details: error,
        });
      }

      const project = await projectResponse.json();

      // Fetch environment variables
      const envResponse = await fetch(`${VERCEL_API_BASE}/v9/projects/${project_id}/env`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const envData = envResponse.ok ? await envResponse.json() : { envs: [] };

      return this.success({
        project: {
          id: project.id,
          name: project.name,
          domain: project.targets?.production?.alias?.[0] || `${project.name}.vercel.app`,
          framework: project.framework,
        },
        build: {
          command: project.buildCommand || 'npm run build',
          output_dir: project.outputDirectory || '.next',
          install_command: project.installCommand,
        },
        env: envData.envs?.map((e: any) => ({
          key: e.key,
          value: e.value,
          target: e.target,
        })) || [],
      });
    } catch (err) {
      return this.error({
        code: 'NETWORK_ERROR',
        message: err instanceof Error ? err.message : String(err),
        recoverable: true,
      });
    }
  }

  async syncEnv(params: SyncEnvParams): Promise<BridgeResponse<SyncEnvData>> {
    const { token, project_id, env_vars } = params;

    try {
      const failed: string[] = [];
      let synced = 0;

      for (const env of env_vars) {
        const response = await fetch(`${VERCEL_API_BASE}/v10/projects/${project_id}/env`, {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            key: env.key,
            value: env.value,
            target: env.target,
            type: 'encrypted',
          }),
        });

        if (response.ok) {
          synced++;
        } else {
          failed.push(env.key);
        }
      }

      return this.success({ synced, failed });
    } catch (err) {
      return this.error({
        code: 'NETWORK_ERROR',
        message: err instanceof Error ? err.message : String(err),
        recoverable: true,
      });
    }
  }

  async deployPreview(params: DeployPreviewParams): Promise<BridgeResponse<DeployPreviewData>> {
    return this.unsupported('deploy:preview');
  }

  async dnsUpdate(params: DnsUpdateParams): Promise<BridgeResponse<DnsUpdateData>> {
    return this.unsupported('dns:update');
  }

  async dnsRollback(params: DnsRollbackParams): Promise<BridgeResponse<DnsRollbackData>> {
    return this.unsupported('dns:rollback');
  }
}

// CLI entry point
const adapter = new VercelAdapter();
const verb = process.argv[2];
let params: unknown;

// Read params from stdin if available
if (process.stdin.isTTY === false) {
  const stdin = await Bun.stdin.text();
  if (stdin.trim()) {
    try {
      params = JSON.parse(stdin);
    } catch (err) {
      console.error('Failed to parse stdin JSON:', err);
      process.exit(1);
    }
  }
}

await adapter.execute(verb, params);
