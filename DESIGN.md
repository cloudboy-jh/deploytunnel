# Deploy Tunnel - Design Document

## Executive Summary

Deploy Tunnel is a CLI-first migration utility designed to facilitate **zero-downtime migrations** between cloud hosting providers. The tool leverages a hybrid architecture (Go + Bun/TypeScript) to provide robust migration workflows with preview tunnels, environment synchronization, route verification, and instant rollback capabilities.

## Design Principles

1. **Safety First**: All migrations should be testable and reversible before committing to DNS changes
2. **Provider Agnostic**: Extensible adapter architecture supports any hosting provider
3. **Developer Experience**: Beautiful, informative CLI with clear progress indicators
4. **Security by Default**: OS keychain integration, no plaintext credentials
5. **State Management**: Full audit trail and migration history
6. **Composable Commands**: Unix philosophy - each command does one thing well

## System Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     CLI Interface (Go)                       │
│  ┌──────────┬──────────┬─────────┬──────────┬─────────┐   │
│  │   init   │   auth   │  tunnel │  verify  │ cutover │   │
│  └──────────┴──────────┴─────────┴──────────┴─────────┘   │
└──────────┬──────────────────────────────────────────┬──────┘
           │                                          │
    ┌──────▼──────┐                            ┌─────▼──────┐
    │   Bridge    │◄─────JSON/Subprocess──────►│  Adapters  │
    │  (Go Core)  │                            │(Bun/TS SDK)│
    └──────┬──────┘                            └────────────┘
           │                                           │
    ┌──────▼──────────────┐                    ┌─────▼─────┐
    │  State Management   │                    │ Provider  │
    │     (SQLite)        │                    │   APIs    │
    │                     │                    │           │
    │ ┌─────────────────┐ │                    │ • Vercel  │
    │ │   Migrations    │ │                    │ • CF      │
    │ │   Env Vars      │ │                    │ • Render  │
    │ │   DNS Records   │ │                    │ • Netlify │
    │ │   Logs          │ │                    └───────────┘
    │ └─────────────────┘ │
    └─────────────────────┘
           │
    ┌──────▼──────────────┐
    │  OS Keychain        │
    │  (Secure Storage)   │
    └─────────────────────┘
```

### Component Breakdown

#### 1. CLI Layer (Go)

**Purpose**: User-facing interface, orchestration, local state management

**Key Modules**:
- `cmd/deploy-tunnel`: Main entry point, command routing
- `internal/cli`: Command implementations (init, auth, tunnel, etc.)
- `ui/`: Lipgloss-based styling and UI components

**Responsibilities**:
- Parse command-line arguments
- Manage local state (SQLite)
- Orchestrate adapter calls via bridge
- Handle keychain operations
- Render progress and results

#### 2. Bridge Layer (Go)

**Purpose**: Facilitate communication between Go core and TypeScript adapters

**Key Modules**:
- `internal/bridge/bridge.go`: Subprocess execution, JSON marshaling
- `internal/bridge/types.go`: Type definitions for all commands

**Protocol**:
- **Transport**: JSON over stdin/stdout
- **Execution**: `bun run adapters/{provider}/index.ts {verb}`
- **Timeout**: 30s default (configurable)
- **Retry**: Exponential backoff for recoverable errors

**Error Handling**:
- Typed error codes (AUTH_FAILED, NETWORK_ERROR, etc.)
- Recoverable vs. non-recoverable distinction
- Detailed error context in `details` field

#### 3. Adapter Layer (Bun + TypeScript)

**Purpose**: Provider-specific API integration

**Key Modules**:
- `adapters/base.ts`: Base adapter class with common logic
- `adapters/types.ts`: TypeScript definitions matching Go types
- `adapters/{provider}/index.ts`: Provider implementations

**Adapter Interface**:
```typescript
interface Adapter {
  capabilities(): Promise<BridgeResponse<CapabilitiesData>>;
  authStart(params: AuthStartParams): Promise<BridgeResponse<AuthStartData>>;
  authRefresh(params: AuthRefreshParams): Promise<BridgeResponse<AuthRefreshData>>;
  fetchConfig(params: FetchConfigParams): Promise<BridgeResponse<FetchConfigData>>;
  syncEnv(params: SyncEnvParams): Promise<BridgeResponse<SyncEnvData>>;
  deployPreview(params: DeployPreviewParams): Promise<BridgeResponse<DeployPreviewData>>;
  dnsUpdate(params: DnsUpdateParams): Promise<BridgeResponse<DnsUpdateData>>;
  dnsRollback(params: DnsRollbackParams): Promise<BridgeResponse<DnsRollbackData>>;
}
```

**Benefits of Bun/TypeScript**:
- Rich ecosystem of provider SDKs (Vercel, Cloudflare, etc.)
- TypeScript type safety for API interactions
- Bun's performance and built-in tools
- Easy community contributions

#### 4. State Management (SQLite)

**Purpose**: Persistent storage of migration state, history, and audit logs

**Schema**:
- `migrations`: Core migration records (id, source, target, domain, status)
- `env_vars`: Environment variable mappings
- `dns_records`: DNS changes with rollback data
- `logs`: Audit trail of all operations

**Location**: `~/.deploy-tunnel/state.db`

**Benefits**:
- No external database required
- ACID transactions
- Queryable history
- Portable file format

#### 5. Keychain Integration

**Purpose**: Secure credential storage using OS-native solutions

**Implementation**: `github.com/zalando/go-keyring`

**Storage Keys**:
- `{provider}-token`: Access tokens
- `{provider}-refresh-token`: OAuth refresh tokens

**Platform Support**:
- macOS: Keychain Access
- Linux: libsecret (GNOME Keyring, KWallet)
- Windows: Credential Manager

## Core Workflows

### 1. Migration Initialization

```
User runs: dt init

┌────────────┐
│ Prompt for│
│ source,   │──► Validate
│ target,   │    providers
│ domain    │
└─────┬──────┘
      │
      ▼
┌────────────────┐
│ Create         │
│ migration      │──► Generate UUID
│ record in DB   │    Store in SQLite
└────────────────┘
      │
      ▼
┌────────────────┐
│ Check auth     │
│ status for     │──► Query keychain
│ providers      │    Display status
└────────────────┘
      │
      ▼
  Show next steps
```

### 2. Authentication Flow

```
User runs: dt auth vercel

┌────────────┐
│ Query      │
│ adapter    │──► bun run vercel/index.ts capabilities
│ capabilities│
└─────┬──────┘
      │
      ▼
┌────────────────┐
│ Start auth     │──► bun run vercel/index.ts auth:start
│ (OAuth or      │    Returns auth_url or prompts for token
│  token flow)   │
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Open browser   │
│ or prompt for  │──► User completes auth
│ token input    │
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Store token    │──► keychain.Store(provider, token)
│ in keychain    │    OS-native secure storage
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Verify token   │──► Test API call
│ with test call │    Ensure token works
└────────────────┘
```

### 3. Environment Sync (Future)

```
User runs: dt sync env

┌────────────────┐
│ Fetch env vars │
│ from source    │──► bridge.FetchConfig(source)
│ provider       │    Returns env array
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Store mappings │
│ in state DB    │──► Save to env_vars table
│                │    Allow key remapping
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Push env vars  │
│ to target      │──► bridge.SyncEnv(target, envVars)
│ provider       │    Batch create/update
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Display        │
│ sync results   │──► Show synced count, failures
└────────────────┘
```

### 4. Tunnel Creation (Future)

**Two Modes**:

#### A. Local Proxy Mode
```go
// Reverse proxy running on localhost:8080
// Routes traffic: source ← localhost:8080 → target

httputil.ReverseProxy{
  Director: func(req *http.Request) {
    // Alternate between source and target
    // Or route based on cookie/header
  }
}
```

#### B. Cloudflare Worker Mode
```javascript
// Deploy worker that proxies between providers
export default {
  async fetch(request) {
    const url = new URL(request.url);
    
    // Check migration status
    if (inMigration(url.hostname)) {
      // Route to target
      return fetch(targetURL, request);
    }
    
    // Default to source
    return fetch(sourceURL, request);
  }
}
```

### 5. Route Verification (Future)

```
User runs: dt verify

┌────────────────┐
│ Discover       │
│ routes         │──► Crawl site, parse sitemap,
│                │    or use provided list
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Fetch routes   │
│ concurrently   │──► GET source.com/route
│ from both      │    GET target.com/route
│ providers      │
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Compare        │
│ responses      │──► Status code
│                │    Headers
│                │    Body hash
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Generate       │
│ report         │──► Table: ✓ ⚠ ✗
│                │    Latency delta
│                │    Fail if >5% mismatch
└────────────────┘
```

### 6. DNS Cutover (Future)

```
User runs: dt cutover

┌────────────────┐
│ Confirm        │
│ action         │──► "Ready to update DNS? (y/n)"
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Fetch current  │
│ DNS records    │──► Store in dns_records table
│                │    Keep for rollback
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Update DNS     │──► bridge.DnsUpdate(...)
│ to point to    │    Change A/CNAME to target
│ target         │
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Wait for       │
│ propagation    │──► Show progress bar
│                │    Check with DNS queries
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Verify target  │
│ is receiving   │──► Test actual domain
│ traffic        │
└────────────────┘
```

### 7. Rollback (Future)

```
User runs: dt rollback

┌────────────────┐
│ Fetch rollback │
│ DNS records    │──► Query dns_records table
│ from state     │    Find previous values
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Restore DNS    │──► bridge.DnsRollback(...)
│ to previous    │    Revert A/CNAME to source
│ values         │
└────────┬───────┘
         │
         ▼
┌────────────────┐
│ Update         │
│ migration      │──► Set status = 'rolled_back'
│ status         │    Log event
└────────────────┘
```

## Error Handling Strategy

### Error Code Hierarchy

```
AUTH_FAILED      → Retry with new token
AUTH_REQUIRED    → Run dt auth <provider>
PROVIDER_ERROR   → Check provider status page
NETWORK_ERROR    → Retry with backoff
INVALID_PARAMS   → Fix command arguments
NOT_FOUND        → Verify resource exists
RATE_LIMITED     → Wait and retry
UNSUPPORTED      → Feature not available for provider
TIMEOUT          → Increase timeout or retry
UNKNOWN          → Log full error details
```

### Retry Policy

```go
type RetryConfig struct {
  MaxAttempts      int           // 3
  InitialBackoff   time.Duration // 1s
  MaxBackoff       time.Duration // 30s
  BackoffMultiplier float64      // 2.0
  RecoverableErrors []ErrorCode  // NETWORK_ERROR, TIMEOUT, RATE_LIMITED
}
```

### User-Facing Errors

```
✗ Authentication failed
  Run: dt auth vercel
  
⚠ Rate limit exceeded (resets in 5m)
  Retrying automatically...
  
✗ Network timeout
  Check your internet connection
  Retrying (2/3)...
```

## Security Considerations

### Threat Model

**Assets**:
- Provider API tokens
- Environment variables (may contain secrets)
- DNS configuration
- Migration state

**Threats**:
1. Token theft from disk
2. Token exposure in logs
3. Man-in-the-middle attacks on adapter communication
4. Unauthorized DNS changes
5. Accidental credential commits to git

**Mitigations**:
1. ✅ OS keychain for token storage (encrypted at rest)
2. ✅ No tokens in logs or database
3. ✅ Local subprocess communication (no network)
4. ⏳ Confirmation prompts for destructive actions
5. ✅ .gitignore for state.db and config files

### Credential Lifecycle

```
┌──────────┐     Store      ┌──────────┐
│   User   │───────────────►│ Keychain │
│  Input   │                │ (OS)     │
└──────────┘                └────┬─────┘
                                 │
                                 │ Get
                                 │
                            ┌────▼─────┐
                            │  Bridge  │
                            │  Call    │
                            └────┬─────┘
                                 │
                                 │ Pass to adapter
                                 │ (subprocess stdin)
                            ┌────▼─────┐
                            │ Adapter  │
                            │  (Bun)   │
                            └──────────┘
                            
Token never touches:
- SQLite database
- Log files
- Disk (except OS keychain)
- Environment variables
```

## Performance Considerations

### Adapter Call Optimization

- **Concurrent Calls**: Use `sync.WaitGroup` for parallel adapter operations
- **Connection Pooling**: Reuse HTTP clients in adapters
- **Caching**: Cache adapter capabilities for session duration
- **Timeouts**: Aggressive timeouts (30s) with retry logic

### Database Optimization

- **Indexes**: On `migration_id`, `status`, `ts` fields
- **Batch Inserts**: Use transactions for bulk env var sync
- **WAL Mode**: Enable for better concurrency
- **Vacuum**: Periodic cleanup of deleted records

### UI Performance

- **Streaming Output**: Use spinners for long-running operations
- **Incremental Updates**: Update progress without redrawing entire UI
- **Async Logging**: Write logs in background goroutine

## Testing Strategy

### Unit Tests

- Go: `go test ./...`
- TypeScript: `bun test`

**Coverage Targets**:
- Bridge protocol: 100%
- State management: 90%
- CLI commands: 80%
- Adapters: 70%

### Integration Tests

```bash
# Test full auth flow
dt auth vercel --token "$TEST_TOKEN"

# Test config fetch
dt fetch:config --project "$TEST_PROJECT_ID"

# Test env sync
dt sync env --dry-run
```

### E2E Tests (Future)

- Spin up test deployments on multiple providers
- Run full migration workflow
- Verify DNS changes
- Test rollback

## Future Enhancements

### Phase 3: Advanced Features

1. **Background Agent**
   - Monitor tunnel health
   - Auto-rollback on errors
   - Metrics collection

2. **Web UI**
   - Built with Bun + React
   - Real-time migration status
   - Visual diff of environments
   - Serve on `localhost:4050`

3. **Plugin Registry**
   - npm packages: `dtu-adapter-aws`, `dtu-adapter-fly`
   - Auto-discovery of installed adapters
   - Version management

4. **Team Features**
   - Shared migrations (via cloud sync)
   - Role-based access control
   - Approval workflows
   - Slack/Discord notifications

5. **Enterprise Mode**
   - SAML/SSO authentication
   - Audit logging to SIEM
   - Compliance reports
   - Multi-region support

## Conclusion

Deploy Tunnel provides a robust, extensible framework for zero-downtime migrations between cloud providers. The hybrid Go + Bun architecture balances performance, security, and developer experience. The adapter pattern ensures easy extensibility while maintaining a consistent user interface across all providers.

The foundational components (CLI, bridge, state management, auth) are now complete and functional. Future phases will add the core migration features (tunnels, verification, cutover) that leverage this solid foundation.

---

**Status**: Phase 1 Complete ✅
**Next**: Implement tunnel engine and DNS management
**Target**: Production-ready by Q2 2025
