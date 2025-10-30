# Deploy Tunnel - Usage Examples

This document provides practical examples of using Deploy Tunnel for common migration scenarios.

## Example 1: Vercel to Cloudflare Migration

### Step 1: Initialize Migration

```bash
$ dt init

╭────────────────────────────────╮
│  DEPLOY ▸ TUNNEL               │
│  migrate safely between hosts  │
╰────────────────────────────────╯

ℹ Let's set up your migration

? Source provider (where you're migrating FROM)

  1) vercel
  2) cloudflare
  3) render
  4) netlify

Enter number: 1

? Target provider (where you're migrating TO)

  1) vercel
  2) cloudflare
  3) render
  4) netlify

Enter number: 2

? Domain name to migrate: myapp.com

ℹ Creating migration configuration...
✓ Migration initialized

Migration ID: 550e8400-e29b-41d4-a716-446655440000
Source: vercel
Target: cloudflare
Domain: myapp.com

ℹ Checking authentication status...

! No credentials found for vercel
ℹ Run: dt auth vercel

! No credentials found for cloudflare
ℹ Run: dt auth cloudflare

ℹ Next steps:
  • Authenticate providers: dt auth vercel && dt auth cloudflare
  • Fetch source configuration: dt fetch:config
  • Sync environment variables: dt sync env
  • Create preview tunnel: dt tunnel create --preview
  • Verify routes: dt verify
  • Cutover when ready: dt cutover
```

### Step 2: Authenticate with Vercel

```bash
$ dt auth vercel

╭────────────────────────────────╮
│  DEPLOY ▸ TUNNEL               │
│  migrate safely between hosts  │
╰────────────────────────────────╯

ℹ Checking vercel adapter capabilities...
✓ Adapter: vercel v1.0.0
Auth Type: token

ℹ Starting authentication...

ℹ This provider requires a personal access token
? Enter your token: ********************************

ℹ Storing credentials securely...
ℹ Verifying credentials...
✓ Authentication successful!

ℹ Your credentials have been securely stored in the system keychain
```

**Where to get Vercel token:**
1. Go to https://vercel.com/account/tokens
2. Click "Create Token"
3. Give it a name (e.g., "deploy-tunnel")
4. Copy the token and paste it when prompted

### Step 3: Authenticate with Cloudflare

```bash
$ dt auth cloudflare

# Similar flow as Vercel
# Token from: https://dash.cloudflare.com/profile/api-tokens
```

### Step 4: List Authenticated Providers

```bash
$ dt auth list

╭────────────────────────────────╮
│  DEPLOY ▸ TUNNEL               │
│  migrate safely between hosts  │
╰────────────────────────────────╯

ℹ Stored credentials:

✓ vercel
✓ cloudflare
```

### Step 5: Fetch Source Configuration (Coming Soon)

```bash
$ dt fetch:config

ℹ Fetching configuration from vercel...

✓ Project found: myapp
Domain: myapp.com
Framework: nextjs

Build Configuration:
  Command: npm run build
  Output: .next
  Install: npm install

Environment Variables: 12 found
  • DATABASE_URL (production)
  • API_KEY (production, preview)
  • NEXT_PUBLIC_API_URL (production, preview, development)
  ...

ℹ Configuration saved to migration state
```

### Step 6: Sync Environment Variables (Coming Soon)

```bash
$ dt sync env

ℹ Syncing environment variables...

Source (vercel): 12 variables
Target (cloudflare): 0 variables

? Confirm sync? (y/n): y

ℹ Pushing variables to cloudflare...
  [████████████████████████████████████████] 100%

✓ Synced: 12 variables
✗ Failed: 0 variables

Note: Sensitive values are encrypted at rest
```

### Step 7: Create Preview Tunnel (Coming Soon)

```bash
$ dt tunnel create --preview

ℹ Creating preview tunnel...

✓ Target deployment created
  URL: https://myapp-preview-abc123.pages.dev

✓ Local proxy started
  Listening on: http://localhost:8080

ℹ Traffic flow:
  localhost:8080 → Target (70%)
  localhost:8080 → Source (30%)

ℹ Test your migration at http://localhost:8080
  Press [r] to rollback | [q] to quit
```

### Step 8: Verify Routes (Coming Soon)

```bash
$ dt verify --routes 100

ℹ Discovering routes...
✓ Found 47 routes from sitemap

ℹ Verifying routes (47 total)...
  [████████████████████████████████████████] 100%

Results:
Route                    Source  Target  Status
───────────────────────  ──────  ──────  ──────
/                        200     200     ✓
/about                   200     200     ✓
/blog                    200     200     ✓
/blog/post-1             200     200     ✓
/api/users               200     200     ✓
/api/posts               200     500     ✗
...

Summary:
  ✓ Passed: 45 (95.7%)
  ! Warnings: 1 (2.1%)
  ✗ Failed: 1 (2.1%)

! /api/posts returned 500 on target
  Investigate before cutover

✓ Verification complete (95.7% pass rate)
```

### Step 9: Cutover DNS (Coming Soon)

```bash
$ dt cutover --ttl 300

╭────────────────────────────────╮
│  🚨 DNS CUTOVER                │
│  This will update DNS records  │
╰────────────────────────────────╯

Domain: myapp.com
Current: vercel (A → 76.76.21.21)
New: cloudflare (CNAME → myapp.pages.dev)
TTL: 300 seconds (5 minutes)

Note: This action is reversible via 'dt rollback'

? Ready to update DNS? (y/n): y

ℹ Backing up current DNS records...
✓ Backup saved to state.db

ℹ Updating DNS records...
✓ A record updated
✓ CNAME record updated

ℹ Waiting for DNS propagation (300s TTL)...
  [████████████────────────────────] 45%

✓ DNS propagation complete
✓ myapp.com now points to cloudflare

ℹ Monitor your application for the next 24 hours
  Rollback available: dt rollback
```

### Step 10: Rollback (if needed)

```bash
$ dt rollback

! Rolling back DNS to previous configuration

Domain: myapp.com
Previous: vercel (A → 76.76.21.21)
Current: cloudflare (CNAME → myapp.pages.dev)

? Confirm rollback? (y/n): y

ℹ Restoring DNS records...
✓ DNS rolled back to vercel

ℹ Migration status: rolled_back
```

---

## Example 2: Render to Vercel Migration

```bash
# Initialize
dt init
# Select: render → vercel

# Authenticate
dt auth render
dt auth vercel

# Run migration
dt fetch:config
dt sync env
dt tunnel create --preview
dt verify
dt cutover
```

---

## Example 3: Testing Adapter Communication

### Check Adapter Capabilities

```bash
$ echo '{}' | bun run adapters/vercel/index.ts capabilities

{
  "ok": true,
  "data": {
    "adapter_name": "vercel",
    "adapter_version": "1.0.0",
    "supported_verbs": [
      "capabilities",
      "auth:start",
      "fetch:config",
      "sync:env",
      "deploy:preview",
      "dns:update",
      "dns:rollback"
    ],
    "auth_type": "token",
    "features": {
      "dns_management": true,
      "preview_deployments": true,
      "env_variables": true,
      "build_logs": true
    }
  },
  "adapter_version": "1.0.0"
}
```

### Test Auth Flow

```bash
$ echo '{"provider":"vercel"}' | bun run adapters/vercel/index.ts auth:start

{
  "ok": true,
  "data": {
    "auth_url": "https://vercel.com/account/tokens"
  },
  "adapter_version": "1.0.0"
}
```

---

## Example 4: Revoking Credentials

```bash
# Revoke specific provider
$ dt auth revoke vercel

╭────────────────────────────────╮
│  DEPLOY ▸ TUNNEL               │
│  migrate safely between hosts  │
╰────────────────────────────────╯

✓ Credentials for vercel have been removed
```

---

## Example 5: CI/CD Integration (Future)

```yaml
# .github/workflows/migrate.yml
name: Deploy Tunnel Migration

on:
  workflow_dispatch:
    inputs:
      action:
        description: 'Action to perform'
        required: true
        type: choice
        options:
          - verify
          - cutover
          - rollback

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Deploy Tunnel
        run: |
          wget https://github.com/johnhorton/deploy-tunnel/releases/latest/download/dt-linux-amd64
          chmod +x dt-linux-amd64
          sudo mv dt-linux-amd64 /usr/local/bin/dt
      
      - name: Authenticate
        env:
          VERCEL_TOKEN: ${{ secrets.VERCEL_TOKEN }}
          CLOUDFLARE_TOKEN: ${{ secrets.CLOUDFLARE_TOKEN }}
        run: |
          echo "$VERCEL_TOKEN" | dt auth vercel
          echo "$CLOUDFLARE_TOKEN" | dt auth cloudflare
      
      - name: Run Action
        run: dt ${{ github.event.inputs.action }} --json
```

---

## Troubleshooting

### "Failed to open database"

```bash
# Check permissions
ls -la ~/.deploy-tunnel/

# Reset state (caution: deletes history)
rm -rf ~/.deploy-tunnel/
dt init
```

### "Adapter not found"

```bash
# Ensure adapters are in the correct location
ls adapters/vercel/index.ts

# Rebuild if necessary
make build
```

### "Authentication failed"

```bash
# Check stored credentials
dt auth list

# Re-authenticate
dt auth revoke vercel
dt auth vercel
```

### "Network timeout"

```bash
# Increase timeout (future feature)
dt fetch:config --timeout 60s

# Check internet connection
curl -I https://api.vercel.com
```

---

## Pro Tips

1. **Always verify before cutover**: Run `dt verify` multiple times to ensure consistency
2. **Use short TTLs**: Start with a 300s (5 min) TTL for faster rollback if needed
3. **Monitor after cutover**: Watch error rates and metrics for 24 hours
4. **Test in preview mode**: Use `--preview` flag to test without affecting production
5. **Keep backups**: Deploy Tunnel stores rollback data, but keep your own backups too

---

For more information, see:
- [README.md](./README.md) - Full documentation
- [DESIGN.md](./DESIGN.md) - Architecture and design decisions
- [bridge_spec.json](./bridge_spec.json) - Adapter protocol specification
