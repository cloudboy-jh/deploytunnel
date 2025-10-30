import type {
  Adapter,
  BridgeResponse,
  BridgeError,
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
} from './types';

export abstract class BaseAdapter implements Adapter {
  protected version = '1.0.0';

  abstract capabilities(): Promise<BridgeResponse<CapabilitiesData>>;
  abstract authStart(params: AuthStartParams): Promise<BridgeResponse<AuthStartData>>;
  abstract authRefresh(params: AuthRefreshParams): Promise<BridgeResponse<AuthRefreshData>>;
  abstract fetchConfig(params: FetchConfigParams): Promise<BridgeResponse<FetchConfigData>>;
  abstract syncEnv(params: SyncEnvParams): Promise<BridgeResponse<SyncEnvData>>;
  abstract deployPreview(params: DeployPreviewParams): Promise<BridgeResponse<DeployPreviewData>>;
  abstract dnsUpdate(params: DnsUpdateParams): Promise<BridgeResponse<DnsUpdateData>>;
  abstract dnsRollback(params: DnsRollbackParams): Promise<BridgeResponse<DnsRollbackData>>;

  protected success<T>(data: T): BridgeResponse<T> {
    return {
      ok: true,
      data,
      adapter_version: this.version,
    };
  }

  protected error(error: BridgeError): BridgeResponse<never> {
    return {
      ok: false,
      error,
      adapter_version: this.version,
    };
  }

  protected unsupported(verb: string): BridgeResponse<never> {
    return this.error({
      code: 'UNSUPPORTED',
      message: `Command '${verb}' is not supported by this adapter`,
      recoverable: false,
    });
  }

  /**
   * Execute adapter command from CLI
   */
  async execute(verb: string, params?: unknown): Promise<void> {
    let response: BridgeResponse<unknown>;

    try {
      switch (verb) {
        case 'capabilities':
          response = await this.capabilities();
          break;
        case 'auth:start':
          response = await this.authStart(params as AuthStartParams);
          break;
        case 'auth:refresh':
          response = await this.authRefresh(params as AuthRefreshParams);
          break;
        case 'fetch:config':
          response = await this.fetchConfig(params as FetchConfigParams);
          break;
        case 'sync:env':
          response = await this.syncEnv(params as SyncEnvParams);
          break;
        case 'deploy:preview':
          response = await this.deployPreview(params as DeployPreviewParams);
          break;
        case 'dns:update':
          response = await this.dnsUpdate(params as DnsUpdateParams);
          break;
        case 'dns:rollback':
          response = await this.dnsRollback(params as DnsRollbackParams);
          break;
        default:
          response = this.error({
            code: 'INVALID_PARAMS',
            message: `Unknown verb: ${verb}`,
            recoverable: false,
          });
      }
    } catch (err) {
      response = this.error({
        code: 'UNKNOWN',
        message: err instanceof Error ? err.message : String(err),
        recoverable: false,
        details: { stack: err instanceof Error ? err.stack : undefined },
      });
    }

    // Write response to stdout
    console.log(JSON.stringify(response));
  }
}
