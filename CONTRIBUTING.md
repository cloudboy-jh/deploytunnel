# Contributing to Deploy Tunnel

Thank you for your interest in contributing to Deploy Tunnel! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.23 or higher
- Bun 1.0 or higher
- Git

### Getting Started

```bash
# Clone the repository
git clone https://github.com/johnhorton/deploy-tunnel.git
cd deploy-tunnel

# Install dependencies
make deps

# Build the binary
make build

# Run tests
make test

# Try it out
./dt help
```

## Project Structure

```
deploy-tunnel/
├── cmd/deploy-tunnel/      # Main CLI entry point
├── internal/               # Go internal packages
│   ├── cli/               # Command implementations
│   ├── bridge/            # Adapter bridge protocol
│   ├── state/             # SQLite state management
│   ├── keychain/          # Secure credential storage
│   ├── tunnel/            # Tunnel engine (TODO)
│   ├── dns/               # DNS management (TODO)
│   └── verify/            # Route verification (TODO)
├── adapters/              # Bun/TypeScript provider adapters
│   ├── vercel/
│   ├── cloudflare/        # (TODO)
│   └── render/            # (TODO)
├── ui/                    # CLI UI components
└── tests/                 # Integration tests
```

## Contributing Code

### 1. Find or Create an Issue

Before starting work, check if an issue exists for what you want to do. If not, create one to discuss your proposed changes.

### 2. Fork and Branch

```bash
# Fork the repo on GitHub, then:
git clone https://github.com/YOUR_USERNAME/deploy-tunnel.git
cd deploy-tunnel

# Create a feature branch
git checkout -b feature/your-feature-name
```

### 3. Make Your Changes

Follow the coding standards below and ensure tests pass.

### 4. Test Your Changes

```bash
# Run all tests
make test

# Run specific Go tests
go test ./internal/bridge/...

# Test adapters
cd adapters && bun test

# Manual testing
./dt init
```

### 5. Commit and Push

```bash
# Format your code
make fmt

# Lint
make lint

# Commit with a clear message
git add .
git commit -m "feat: add cloudflare adapter"

# Push to your fork
git push origin feature/your-feature-name
```

### 6. Create a Pull Request

Open a PR on GitHub with:
- Clear description of changes
- Reference to related issue(s)
- Screenshots/output if relevant
- Test results

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting (run `make fmt`)
- Write meaningful comments for exported functions
- Keep functions small and focused
- Use meaningful variable names

**Example:**
```go
// FetchConfig retrieves project configuration from the provider
func (b *Bridge) FetchConfig(ctx context.Context, params FetchConfigParams) (*FetchConfigData, error) {
    resp, err := b.Execute(ctx, params.Provider, "fetch:config", params)
    if err != nil {
        return nil, err
    }
    
    var data FetchConfigData
    if err := mapToStruct(resp.Data, &data); err != nil {
        return nil, fmt.Errorf("failed to parse config data: %w", err)
    }
    
    return &data, nil
}
```

### TypeScript Style

- Use TypeScript strict mode
- Define proper types (avoid `any`)
- Use async/await (not callbacks)
- Export types from `types.ts`

**Example:**
```typescript
async fetchConfig(params: FetchConfigParams): Promise<BridgeResponse<FetchConfigData>> {
  try {
    const response = await fetch(`${API_BASE}/projects/${params.project_id}`, {
      headers: { Authorization: `Bearer ${params.token}` },
    });
    
    if (!response.ok) {
      return this.error({
        code: 'PROVIDER_ERROR',
        message: 'Failed to fetch config',
        recoverable: false,
      });
    }
    
    const data = await response.json();
    return this.success(data);
  } catch (err) {
    return this.error({
      code: 'NETWORK_ERROR',
      message: err instanceof Error ? err.message : String(err),
      recoverable: true,
    });
  }
}
```

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add cloudflare adapter
fix: handle timeout errors in bridge
docs: update README with installation steps
test: add tests for state management
refactor: simplify auth flow
chore: update dependencies
```

## Creating a New Provider Adapter

### 1. Create Adapter Directory

```bash
mkdir -p adapters/my-provider
```

### 2. Implement Adapter

```typescript
// adapters/my-provider/index.ts
#!/usr/bin/env bun

import { BaseAdapter } from '../base';
import type {
  BridgeResponse,
  CapabilitiesData,
  AuthStartParams,
  AuthStartData,
  FetchConfigParams,
  FetchConfigData,
  // ... other types
} from '../types';

class MyProviderAdapter extends BaseAdapter {
  async capabilities(): Promise<BridgeResponse<CapabilitiesData>> {
    return this.success({
      adapter_name: 'my-provider',
      adapter_version: '1.0.0',
      supported_verbs: ['auth:start', 'fetch:config'],
      auth_type: 'oauth', // or 'token' or 'api_key'
      features: {
        dns_management: true,
        preview_deployments: true,
        env_variables: true,
        build_logs: false,
      },
    });
  }

  async authStart(params: AuthStartParams): Promise<BridgeResponse<AuthStartData>> {
    // Implement OAuth flow or token prompt
    return this.success({
      auth_url: 'https://my-provider.com/oauth/authorize',
    });
  }

  async fetchConfig(params: FetchConfigParams): Promise<BridgeResponse<FetchConfigData>> {
    // Fetch project configuration from provider API
    try {
      const response = await fetch(`https://api.my-provider.com/projects/${params.project_id}`, {
        headers: { Authorization: `Bearer ${params.token}` },
      });
      
      const data = await response.json();
      
      return this.success({
        project: {
          id: data.id,
          name: data.name,
          domain: data.domain,
          framework: data.framework,
        },
        build: {
          command: data.build_command,
          output_dir: data.output_directory,
        },
        env: data.env_vars,
      });
    } catch (err) {
      return this.error({
        code: 'NETWORK_ERROR',
        message: err instanceof Error ? err.message : String(err),
        recoverable: true,
      });
    }
  }
  
  // Implement other methods...
  // For unsupported features, use: return this.unsupported('verb-name');
}

// CLI entry point
const adapter = new MyProviderAdapter();
const verb = process.argv[2];
let params: unknown;

if (process.stdin.isTTY === false) {
  const stdin = await Bun.stdin.text();
  if (stdin.trim()) {
    params = JSON.parse(stdin);
  }
}

await adapter.execute(verb, params);
```

### 3. Test Your Adapter

```bash
# Test capabilities
echo '{}' | bun run adapters/my-provider/index.ts capabilities

# Test auth
echo '{"provider":"my-provider"}' | bun run adapters/my-provider/index.ts auth:start

# Test with real credentials (use test account)
echo '{"provider":"my-provider","token":"test-token"}' | \
  bun run adapters/my-provider/index.ts fetch:config
```

### 4. Add to CLI

Update `cmd/deploy-tunnel/main.go` to add your provider to the list:

```go
const (
    ProviderVercel     Provider = "vercel"
    ProviderCloudflare Provider = "cloudflare"
    ProviderRender     Provider = "render"
    ProviderMyProvider Provider = "my-provider"  // Add this
)
```

### 5. Document Your Adapter

Add provider-specific documentation in `adapters/my-provider/README.md`:

```markdown
# My Provider Adapter

## Authentication

1. Go to https://my-provider.com/settings/api
2. Create a new API token
3. Run: `dt auth my-provider`
4. Paste the token when prompted

## Features

- [x] Environment variable sync
- [x] DNS management
- [x] Preview deployments
- [ ] Build logs (not supported by API)

## Known Limitations

- API rate limit: 100 requests/minute
- Maximum env vars: 100
- DNS propagation time: ~5 minutes
```

## Testing Guidelines

### Unit Tests

Write tests for all public functions:

```go
// internal/bridge/bridge_test.go
func TestBridgeExecute(t *testing.T) {
    bridge := NewBridge("./testdata/adapters")
    
    resp, err := bridge.Execute(
        context.Background(),
        ProviderVercel,
        "capabilities",
        nil,
    )
    
    assert.NoError(t, err)
    assert.True(t, resp.OK)
}
```

### Integration Tests

Test full workflows:

```bash
#!/bin/bash
# tests/integration/test_auth_flow.sh

# Setup
export TEST_TOKEN="test-token-12345"

# Test
./dt auth vercel --token "$TEST_TOKEN"

# Verify
if ./dt auth list | grep -q "vercel"; then
    echo "✓ Auth test passed"
else
    echo "✗ Auth test failed"
    exit 1
fi

# Cleanup
./dt auth revoke vercel
```

## Documentation

- **Code comments**: Document exported functions and complex logic
- **README.md**: Keep examples up to date
- **DESIGN.md**: Update architecture docs for major changes
- **EXAMPLE.md**: Add usage examples for new features

## Pull Request Checklist

Before submitting your PR, ensure:

- [ ] Code follows project style guidelines
- [ ] Tests are passing (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] Code is linted (`make lint`)
- [ ] New features have tests
- [ ] Documentation is updated
- [ ] Commit messages follow conventional commits
- [ ] PR description is clear and complete

## Getting Help

- **GitHub Issues**: Ask questions or report bugs
- **Discussions**: General questions and ideas
- **Discord**: Real-time chat (coming soon)

## Code of Conduct

Be respectful, inclusive, and constructive. We're all here to build something great together!

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Deploy Tunnel!
