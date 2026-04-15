# BridgeFT MCP Server

MCP (Model Context Protocol) server for the [BridgeFT WealthTech API](https://docs.bridgeft.com). Gives Claude Code direct access to BridgeFT data — accounts, positions, transactions, balances, performance, AUM, securities, households, billing, and more.

## Setup

### 1. Build

```bash
cd ~/workspace/bridgeft
go build -o bridgeft-mcp .
```

### 2. Configure Claude Code

Add to your `~/.claude/settings.json` (or project `.claude/settings.json`):

**Without authentication** (sandbox/internal environments):

```json
{
  "mcpServers": {
    "bridgeft": {
      "command": "~/workspace/bridgeft/bridgeft-mcp"
    }
  }
}
```

**With authentication** (production/external environments):

```json
{
  "mcpServers": {
    "bridgeft": {
      "command": "~/workspace/bridgeft/bridgeft-mcp",
      "env": {
        "BRIDGEFT_CLIENT_ID": "your-client-id",
        "BRIDGEFT_CLIENT_SECRET": "your-client-secret"
      }
    }
  }
}
```

**With custom base URL:**

```json
{
  "mcpServers": {
    "bridgeft": {
      "command": "~/workspace/bridgeft/bridgeft-mcp",
      "env": {
        "BRIDGEFT_BASE_URL": "https://your-custom-endpoint.com/v2",
        "BRIDGEFT_CLIENT_ID": "your-client-id",
        "BRIDGEFT_CLIENT_SECRET": "your-client-secret"
      }
    }
  }
}
```

### 3. Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BRIDGEFT_BASE_URL` | No | `https://api.bridgeft.com/v2` | API base URL |
| `BRIDGEFT_CLIENT_ID` | No | — | OAuth2 client ID (enables auth) |
| `BRIDGEFT_CLIENT_SECRET` | No | — | OAuth2 client secret (enables auth) |

When `BRIDGEFT_CLIENT_ID` and `BRIDGEFT_CLIENT_SECRET` are both set, the server authenticates using OAuth2 client_credentials flow (Basic Auth → Bearer JWT) with automatic token caching and refresh. When unset, requests are made without authentication.

### 4. Enable/Disable

The server doesn't need to run all the time. Options to control it:

- **`/mcp`** — Toggle on/off interactively per session in Claude Code
- **Project-level** — Add to `.claude/settings.local.json` in specific projects instead of global settings
- **CLI flag** — `claude --mcp-server bridgeft=off` to start with it disabled

## Architecture

```
┌─────────────┐       stdio        ┌──────────────────┐      HTTPS/REST     ┌─────────────────┐
│             │   (JSON-RPC 2.0)   │                  │    (Bearer JWT)     │                 │
│ Claude Code │◄──────────────────►│  bridgeft-mcp    │◄───────────────────►│  BridgeFT API   │
│             │                    │  (Go binary)     │                     │  api.bridgeft   │
└─────────────┘                    └──────────────────┘                     │  .com/v2        │
                                          │                                └─────────────────┘
                                          │
                                   ┌──────┴──────┐
                                   │             │
                              ┌────▼────┐  ┌─────▼─────┐
                              │ main.go │  │ client.go  │
                              │         │  │            │
                              │ Server  │  │ HTTP Client│
                              │ 55 tools│  │ OAuth2 auth│
                              │ routing │  │ token cache│
                              └─────────┘  └────────────┘
```

### How it works

1. **Claude Code** spawns `bridgeft-mcp` as a child process and communicates via stdin/stdout using JSON-RPC 2.0 (MCP protocol)
2. **main.go** registers 55 tools on an `mcp.Server` using the [official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk). Tools are built from reusable handler factories (`listHandler`, `getByIDHandler`, `filterHandler`, `latestHandler`) that map MCP tool calls to REST API requests
3. **client.go** handles HTTP communication with the BridgeFT API:
   - If credentials are configured: authenticates via OAuth2 `client_credentials` grant (POST `/oauth2/token` with Basic Auth), caches the JWT, and auto-refreshes 60s before expiry
   - Attaches `Bearer` token to all subsequent requests
   - Handles pagination via `page` and `limit` query parameters

### File structure

```
~/workspace/bridgeft/
├── main.go                 # MCP server, tool definitions, handler factories
├── client.go               # BridgeFT HTTP client with OAuth2 auth
├── go.mod / go.sum         # Go module dependencies
├── bridgeft-mcp            # Compiled binary
├── wealthtech-api.json     # BridgeFT OpenAPI spec (151 endpoints, 109 schemas)
└── README.md
```

### Tool pattern

Most tools follow one of four patterns, built from generic handler factories:

| Pattern | HTTP | Input | Example |
|---------|------|-------|---------|
| `listHandler(path)` | `GET /path?page=&limit=` | `ListInput{Page, Limit}` | `list_accounts` |
| `getByIDHandler(path)` | `GET /path/{id}` | `GetByIDInput{ID}` | `get_account` |
| `filterHandler(path)` | `POST /path/filter` | `FilterInput{Filter, Page, Limit}` | `filter_accounts` |
| `latestHandler(path)` | `GET /path/latest` | `ListInput{Page, Limit}` | `latest_positions` |

Custom handlers are used for analytics (performance, AUM), securities search, and the generic `api_request` tool.

## Available Tools (55 tools)

### Accounts
- `list_accounts` — List all accounts (paginated)
- `get_account` — Get account by ID
- `filter_accounts` — Filter accounts by criteria

### Related Persons
- `list_related_persons` — List account holders, beneficiaries, trustees
- `filter_related_persons` — Filter related persons

### Source Data (Custodial)
- `list_positions` / `get_position` / `filter_positions` / `latest_positions`
- `list_transactions` / `get_transaction` / `filter_transactions` / `latest_transactions`
- `list_balances` / `get_balance` / `filter_balances` / `latest_balances`
- `list_lots` / `get_lot` / `filter_lots` / `latest_lots`
- `list_realized_gains` / `filter_realized_gains` / `latest_realized_gains`

### Securities
- `list_securities` — List all securities
- `get_security` — Get security by ID
- `filter_securities` — Filter securities
- `search_securities` — Search by ticker/name

### Analytics
- `account_performance` — Account returns (TWR/MWR)
- `household_performance` — Household-level performance
- `benchmark_performance` — Benchmark comparison data
- `list_aum` — Firm-wide AUM
- `aum_by_account` — AUM by account
- `aum_by_household` — AUM by household

### Reporting
- `list_households` / `get_household` / `filter_households` / `create_household` / `remap_households`
- `list_benchmarks` / `get_benchmark` / `filter_benchmarks`
- `list_asset_classifications` / `filter_asset_classifications`
- `list_target_allocations` / `filter_target_allocations`
- `list_printable_reports` / `filter_printable_reports`
- `run_web_report`

### Investment Management
- `list_models` / `get_model`
- `list_strategies` / `get_strategy`

### Billing
- `list_billing_groups` / `list_fee_structures`
- `list_invoices` / `filter_invoices`
- `list_billing_reports`

### Infrastructure
- `list_jobs` / `get_job` — Background job status
- `account_summary_status` — Data feed health
- `source_data_status` — Processing pipeline status
- `list_indexes` / `filter_indexes` — IDC market indexes
- `list_advisor_codes` — Firm advisor/rep codes

### Generic
- `api_request` — Raw request to any endpoint (method, path, body, query)

## Usage Examples

Once configured, ask Claude Code:

```
"List my BridgeFT accounts"
"Show latest positions for account 12345"
"What's the performance of account 67890 from 2025-01-01 to 2025-12-31?"
"Search for securities matching 'AAPL'"
"Filter transactions for account 12345 with type 'BUY'"
```
