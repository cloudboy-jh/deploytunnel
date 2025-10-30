# Deploy Tunnel - TUI Guide

Deploy Tunnel is now a **fully interactive TUI application**. Instead of running separate commands, you navigate through beautiful menus and wizards.

## Quick Start

```bash
# Just run dt - no arguments needed!
dt
```

This launches the **main dashboard** where you can:
- Start a new migration
- View migration history
- Manage authentication
- Continue active migrations

## Main Dashboard

When you run `dt`, you see:

```
╭────────────────────────────────╮
│  DEPLOY ▸ TUNNEL               │
│  migrate safely between hosts  │
╰────────────────────────────────╯

┌─────────────────────────────┐
│ Active Migration            │
│                             │
│ Domain:  myapp.com          │
│ Source:  vercel             │
│ Target:  cloudflare         │
│ Status:  pending            │
└─────────────────────────────┘

Main Menu

> Start New Migration
  View Migrations
  Manage Auth
  Current Migration
  Exit

Deploy Tunnel v1.0 | ↑↓ navigate • enter select • q quit
```

**Navigation:**
- Use `↑` and `↓` arrow keys to navigate
- Press `Enter` to select
- Press `q` to quit

## Migration Wizard (dt init)

The init flow is now a **step-by-step wizard**:

### Step 1: Select Source Provider

```
● ● ○ ○
Where are you migrating FROM?

> Vercel
  Cloudflare
  Render
  Netlify
```

### Step 2: Select Target Provider

```
● ● ○ ○
Where are you migrating TO?

✓ Source: vercel

> Vercel
  Cloudflare
  Render
  Netlify
```

### Step 3: Enter Domain

```
● ● ● ○
What domain are you migrating?

✓ Source: vercel
✓ Target: cloudflare

Domain name:
► myapp.com_

Press Enter to continue
```

### Step 4: Confirm

```
● ● ● ●
Confirm Migration Setup

┌──────────────────────────┐
│ Migration Summary        │
│                          │
│ Source:     vercel       │
│             ✓ Authenticated
│                          │
│ Target:     cloudflare   │
│             ✗ Not authenticated
│                          │
│ Domain:     myapp.com    │
└──────────────────────────┘

Press Enter to create migration • q to cancel
```

### Step 5: Complete

```
✓ Migration initialized successfully!

┌──────────────────────────────────────┐
│ Migration ID:                        │
│ 550e8400-e29b-41d4-a716-446655440000 │
│                                      │
│ Next Steps:                          │
│   1. Run 'dt' to open the dashboard  │
│   2. Authenticate with your providers│
│   3. Start the migration workflow    │
└──────────────────────────────────────┘
```

## Authentication (dt auth)

The auth flow is fully interactive:

### Main Auth Menu

```
╭────────────────────────────────╮
│  DEPLOY ▸ TUNNEL               │
│  migrate safely between hosts  │
╰────────────────────────────────╯

Authentication Menu

> Authenticate Provider
  List Authenticated
  Revoke Credentials
  Back to Dashboard
```

### Select Provider

```
Select provider to authenticate:

> Vercel
  ✓ Cloudflare (authenticated)
  Render
  Netlify
```

### Enter Token

```
✓ Adapter: vercel v1.0.0
Auth Type: token

Get your token:
https://vercel.com/account/tokens

Opening in browser...

Paste your token:
► ••••••••••••••••••••••••••••_

Press Enter to continue • Token will be stored securely in your system keychain
```

### Verifying

```
⠋ Verifying credentials...
```

### Success

```
✓ Successfully authenticated with vercel!

Press q to return to dashboard
```

## Keyboard Shortcuts

### Global
- `q` - Quit / Go back
- `Ctrl+C` - Force quit

### Navigation
- `↑` / `↓` - Move up/down in lists
- `Enter` - Select item
- `Tab` - Next field (in forms)
- `Shift+Tab` - Previous field (in forms)

### Text Input
- Type normally
- `Backspace` - Delete character
- `Ctrl+W` - Delete word
- `Ctrl+U` - Clear line

## Command Line Options

While the TUI is the primary interface, you can still launch specific screens:

```bash
dt              # Launch main dashboard (default)
dt init         # Launch migration wizard
dt auth         # Launch auth menu
dt help         # Show help text (non-interactive)
dt version      # Show version (non-interactive)
```

## Tips

1. **Just run `dt`** - Everything is accessible from the main dashboard
2. **Use arrow keys** - All navigation is keyboard-based for speed
3. **Press `q` to go back** - You can always return to the previous screen
4. **Secure by default** - All tokens are stored in your OS keychain
5. **Step indicators** - Progress dots (●) show where you are in multi-step flows

## Workflow Example

### Complete Migration Flow

```bash
# 1. Launch dashboard
dt

# 2. Select "Start New Migration" with Enter
# 3. Follow the wizard:
#    - Select source: Vercel
#    - Select target: Cloudflare  
#    - Enter domain: myapp.com
#    - Confirm

# 4. Back at dashboard, select "Manage Auth"
# 5. Authenticate Vercel
# 6. Authenticate Cloudflare

# 7. Select "Current Migration" from dashboard
# 8. Follow migration workflow:
#    - Fetch config
#    - Sync environment variables
#    - Create preview tunnel
#    - Verify routes
#    - Cutover DNS
```

## Design Principles

The TUI is designed around these principles:

1. **Discoverability** - All options are visible, no hidden commands
2. **Progressive Disclosure** - Information appears when needed
3. **Forgiving** - Easy to go back and change selections
4. **Visual Feedback** - Progress indicators, status colors, confirmations
5. **Keyboard-First** - Optimized for keyboard navigation

## Color Scheme

- **Coral (#ef9f76)** - Primary actions, headers, selected items
- **Green (#a6d189)** - Success states, completed steps
- **Red (#e78284)** - Errors, warnings, destructive actions
- **Yellow (#e5c890)** - Pending states, in-progress items
- **Light Gray (#a5adce)** - Secondary text, descriptions

## Coming Soon

Future TUI screens:

- **Migration Workflow** - Interactive steps for running migrations
- **Route Verification** - Real-time route comparison table
- **Tunnel Monitor** - Live traffic visualization
- **Environment Editor** - Interactive env var mapping
- **Migration History** - Browse past migrations with details

---

**Try it now**: `dt`
