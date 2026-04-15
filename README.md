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

```json
{
  "mcpServers": {
    "bridgeft": {
      "command": "/Users/<username>/<workspace>/bridgeft/bridgeft-mcp"
    }
  }
}
```

Optionally override the base URL:

```json
{
  "mcpServers": {
    "bridgeft": {
      "command": "/Users/<username>/<workspace>/bridgeft/bridgeft-mcp",
      "env": {
        "BRIDGEFT_BASE_URL": "https://your-custom-endpoint.com/v2"
      }
    }
  }
}
```

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

## Architecture

```
Claude Code ←(stdio/JSON-RPC)→ bridgeft-mcp ←(HTTPS/REST)→ BridgeFT API
```

- Transport: stdio (stdin/stdout JSON-RPC per MCP spec)
- Built with: [official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk)
