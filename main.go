package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var apiClient *Client

func main() {
	baseURL := envOrDefault("BRIDGEFT_BASE_URL", "https://api.bridgeft.com/v2")
	clientID := os.Getenv("BRIDGEFT_CLIENT_ID")
	clientSecret := os.Getenv("BRIDGEFT_CLIENT_SECRET")
	apiClient = NewClient(baseURL, clientID, clientSecret)

	server := mcp.NewServer(
		&mcp.Implementation{Name: "bridgeft", Version: "1.0.0"},
		nil,
	)

	registerTools(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
	fmt.Println("MCP Server is running")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// --- Tool Input Types ---

type ListInput struct {
	Page  int `json:"page,omitempty" jsonschema:"page number (default 1)"`
	Limit int `json:"limit,omitempty" jsonschema:"results per page (default 100, max 10000)"`
}

type GetByIDInput struct {
	ID int `json:"id" jsonschema:"the resource ID"`
}

type FilterInput struct {
	Filter map[string]interface{} `json:"filter" jsonschema:"filter criteria as key-value pairs matching the resource fields"`
	Page   int                    `json:"page,omitempty" jsonschema:"page number"`
	Limit  int                    `json:"limit,omitempty" jsonschema:"results per page"`
}

type SearchSecuritiesInput struct {
	Query string `json:"query" jsonschema:"search query (ticker symbol or name)"`
}

type AccountPerformanceInput struct {
	AccountIDs []int  `json:"account_ids" jsonschema:"list of account IDs to get performance for"`
	StartDate  string `json:"start_date,omitempty" jsonschema:"start date (YYYY-MM-DD)"`
	EndDate    string `json:"end_date,omitempty" jsonschema:"end date (YYYY-MM-DD)"`
}

type HouseholdPerformanceInput struct {
	HouseholdIDs []int  `json:"household_ids" jsonschema:"list of household IDs"`
	StartDate    string `json:"start_date,omitempty" jsonschema:"start date (YYYY-MM-DD)"`
	EndDate      string `json:"end_date,omitempty" jsonschema:"end date (YYYY-MM-DD)"`
}

type BenchmarkPerformanceInput struct {
	BenchmarkIDs []int  `json:"benchmark_ids" jsonschema:"list of benchmark IDs"`
	StartDate    string `json:"start_date,omitempty" jsonschema:"start date (YYYY-MM-DD)"`
	EndDate      string `json:"end_date,omitempty" jsonschema:"end date (YYYY-MM-DD)"`
}

type AUMByAccountInput struct {
	AccountIDs []int  `json:"account_ids" jsonschema:"list of account IDs"`
	StartDate  string `json:"start_date,omitempty" jsonschema:"start date (YYYY-MM-DD)"`
	EndDate    string `json:"end_date,omitempty" jsonschema:"end date (YYYY-MM-DD)"`
}

type AUMByHouseholdInput struct {
	HouseholdIDs []int  `json:"household_ids" jsonschema:"list of household IDs"`
	StartDate    string `json:"start_date,omitempty" jsonschema:"start date (YYYY-MM-DD)"`
	EndDate      string `json:"end_date,omitempty" jsonschema:"end date (YYYY-MM-DD)"`
}

type WebReportInput struct {
	ReportType string                 `json:"report_type" jsonschema:"type of web report to run"`
	Params     map[string]interface{} `json:"params" jsonschema:"report parameters"`
}

type APIRequestInput struct {
	Method string                 `json:"method" jsonschema:"HTTP method: GET, POST, PUT, DELETE"`
	Path   string                 `json:"path" jsonschema:"API path e.g. /account-management/accounts"`
	Body   map[string]interface{} `json:"body,omitempty" jsonschema:"request body (for POST/PUT)"`
	Query  map[string]string      `json:"query,omitempty" jsonschema:"query parameters"`
}

type UpdateInput struct {
	ID   int                    `json:"id" jsonschema:"the resource ID to update"`
	Data map[string]interface{} `json:"data" jsonschema:"fields to update"`
}

type CreateInput struct {
	Data map[string]interface{} `json:"data" jsonschema:"resource fields"`
}

type HouseholdRemapInput struct {
	Assignments map[string]interface{} `json:"assignments" jsonschema:"account-to-household assignment mapping"`
}

// --- Helper functions ---

func queryFromPagination(page, limit int) map[string]string {
	q := map[string]string{}
	if page > 0 {
		q["page"] = strconv.Itoa(page)
	}
	if limit > 0 {
		q["limit"] = strconv.Itoa(limit)
	}
	return q
}

func textResult(data json.RawMessage) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}
}

func errResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
		IsError: true,
	}
}

// --- Generic tool handler builders ---

func listHandler(path string) func(context.Context, *mcp.CallToolRequest, ListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListInput) (*mcp.CallToolResult, any, error) {
		data, err := apiClient.Request("GET", path, nil, queryFromPagination(input.Page, input.Limit))
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

func getByIDHandler(pathPrefix string) func(context.Context, *mcp.CallToolRequest, GetByIDInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetByIDInput) (*mcp.CallToolResult, any, error) {
		path := fmt.Sprintf("%s/%d", pathPrefix, input.ID)
		data, err := apiClient.Request("GET", path, nil, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

func filterHandler(path string) func(context.Context, *mcp.CallToolRequest, FilterInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input FilterInput) (*mcp.CallToolResult, any, error) {
		data, err := apiClient.Request("POST", path, input.Filter, queryFromPagination(input.Page, input.Limit))
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

func latestHandler(path string) func(context.Context, *mcp.CallToolRequest, ListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListInput) (*mcp.CallToolResult, any, error) {
		data, err := apiClient.Request("GET", path, nil, queryFromPagination(input.Page, input.Limit))
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	}
}

// --- Tool registration ---

func registerTools(server *mcp.Server) {
	// ==================== ACCOUNTS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_accounts",
		Description: "List all investment accounts accessible to your application. Returns paginated results with account details including name, custodian, registration type, balances, and household assignments.",
	}, listHandler("/account-management/accounts"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_account",
		Description: "Retrieve a single account by its ID. Returns full account details including balances, fee structures, benchmarks, and custodian information.",
	}, getByIDHandler("/account-management/accounts"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_accounts",
		Description: "Filter accounts by criteria. Filterable fields include: id, name, firm_id, custodian_id, household_id, status, account_number, registration_type, etc.",
	}, filterHandler("/account-management/accounts/filter"))

	// ==================== RELATED PERSONS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_related_persons",
		Description: "List all related persons (account holders, beneficiaries, trustees) associated with accounts.",
	}, listHandler("/account-management/related-persons"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_related_persons",
		Description: "Filter related persons by criteria such as account_id, relationship_type, name, email.",
	}, filterHandler("/account-management/related-persons/filter"))

	// ==================== SOURCE POSITIONS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_positions",
		Description: "List all custodial position records. Positions show current holdings in each account with quantities, market values, and cost basis.",
	}, listHandler("/data/source/positions"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_position",
		Description: "Retrieve a single position record by ID.",
	}, getByIDHandler("/data/source/positions"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_positions",
		Description: "Filter position records by criteria such as account_id, security_id, cusip, ticker.",
	}, filterHandler("/data/source/positions/filter"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "latest_positions",
		Description: "Get the latest position records (most recent data feed).",
	}, latestHandler("/data/source/positions/latest"))

	// ==================== SOURCE TRANSACTIONS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_transactions",
		Description: "List all custodial transaction records. Includes buys, sells, dividends, fees, transfers, and other activity.",
	}, listHandler("/data/source/transactions"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_transaction",
		Description: "Retrieve a single transaction record by ID.",
	}, getByIDHandler("/data/source/transactions"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_transactions",
		Description: "Filter transactions by criteria such as account_id, security_id, transaction_type, trade_date, settlement_date.",
	}, filterHandler("/data/source/transactions/filter"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "latest_transactions",
		Description: "Get the latest transaction records from the most recent data feed.",
	}, latestHandler("/data/source/transactions/latest"))

	// ==================== SOURCE ACCOUNT BALANCES ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_balances",
		Description: "List all account balance records showing total market values, cash balances, and accrued interest.",
	}, listHandler("/data/source/account-balances"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_balance",
		Description: "Retrieve a single account balance record by ID.",
	}, getByIDHandler("/data/source/account-balances"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_balances",
		Description: "Filter account balance records by criteria such as account_id, as_of_date.",
	}, filterHandler("/data/source/account-balances/filter"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "latest_balances",
		Description: "Get the latest account balance records.",
	}, latestHandler("/data/source/account-balances/latest"))

	// ==================== SOURCE LOTS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_lots",
		Description: "List all tax lot records showing individual purchase lots with acquisition dates, cost basis, and current values.",
	}, listHandler("/data/source/lots"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_lot",
		Description: "Retrieve a single lot record by ID.",
	}, getByIDHandler("/data/source/lots"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_lots",
		Description: "Filter lot records by criteria such as account_id, security_id, acquisition_date.",
	}, filterHandler("/data/source/lots/filter"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "latest_lots",
		Description: "Get the latest lot records.",
	}, latestHandler("/data/source/lots/latest"))

	// ==================== SOURCE REALIZED GAIN/LOSS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_realized_gains",
		Description: "List all realized gain/loss records from closed positions.",
	}, listHandler("/data/source/realized-gain-loss"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_realized_gains",
		Description: "Filter realized gain/loss records by criteria such as account_id, security_id, close_date.",
	}, filterHandler("/data/source/realized-gain-loss/filter"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "latest_realized_gains",
		Description: "Get the latest realized gain/loss records.",
	}, latestHandler("/data/source/realized-gain-loss/latest"))

	// ==================== SECURITIES ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_securities",
		Description: "List all securities in the system with ticker, CUSIP, name, asset class, and pricing data.",
	}, listHandler("/data/custodian/securities"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_security",
		Description: "Retrieve a single security by ID.",
	}, getByIDHandler("/data/custodian/securities"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_securities",
		Description: "Filter securities by criteria such as ticker, cusip, asset_class, name.",
	}, filterHandler("/data/custodian/securities/filter"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_securities",
		Description: "Search for securities by ticker symbol or name.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchSecuritiesInput) (*mcp.CallToolResult, any, error) {
		path := fmt.Sprintf("/data/custodian/securities/search/%s", input.Query)
		data, err := apiClient.Request("GET", path, nil, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	// ==================== ANALYTICS - PERFORMANCE ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "account_performance",
		Description: "Get performance data (returns, TWR, MWR) for one or more accounts over a date range.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AccountPerformanceInput) (*mcp.CallToolResult, any, error) {
		body := map[string]interface{}{"account_ids": input.AccountIDs}
		if input.StartDate != "" {
			body["start_date"] = input.StartDate
		}
		if input.EndDate != "" {
			body["end_date"] = input.EndDate
		}
		data, err := apiClient.Request("POST", "/analytics/account-performance", body, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "household_performance",
		Description: "Get performance data for one or more households over a date range.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input HouseholdPerformanceInput) (*mcp.CallToolResult, any, error) {
		body := map[string]interface{}{"household_ids": input.HouseholdIDs}
		if input.StartDate != "" {
			body["start_date"] = input.StartDate
		}
		if input.EndDate != "" {
			body["end_date"] = input.EndDate
		}
		data, err := apiClient.Request("POST", "/analytics/household-performance", body, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "benchmark_performance",
		Description: "Get performance data for one or more benchmarks over a date range.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input BenchmarkPerformanceInput) (*mcp.CallToolResult, any, error) {
		body := map[string]interface{}{"benchmark_ids": input.BenchmarkIDs}
		if input.StartDate != "" {
			body["start_date"] = input.StartDate
		}
		if input.EndDate != "" {
			body["end_date"] = input.EndDate
		}
		data, err := apiClient.Request("POST", "/analytics/benchmark-performance", body, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	// ==================== ANALYTICS - AUM ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_aum",
		Description: "List firm-wide Assets Under Management (AUM) records.",
	}, listHandler("/analytics/aum"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "aum_by_account",
		Description: "Get AUM broken down by specific accounts.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AUMByAccountInput) (*mcp.CallToolResult, any, error) {
		body := map[string]interface{}{"account_ids": input.AccountIDs}
		if input.StartDate != "" {
			body["start_date"] = input.StartDate
		}
		if input.EndDate != "" {
			body["end_date"] = input.EndDate
		}
		data, err := apiClient.Request("POST", "/analytics/aum/by-account", body, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "aum_by_household",
		Description: "Get AUM broken down by specific households.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AUMByHouseholdInput) (*mcp.CallToolResult, any, error) {
		body := map[string]interface{}{"household_ids": input.HouseholdIDs}
		if input.StartDate != "" {
			body["start_date"] = input.StartDate
		}
		if input.EndDate != "" {
			body["end_date"] = input.EndDate
		}
		data, err := apiClient.Request("POST", "/analytics/aum/by-household", body, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	// ==================== HOUSEHOLDS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_households",
		Description: "List all households. Households group related accounts (e.g. family members) for consolidated reporting.",
	}, listHandler("/reporting/households"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_household",
		Description: "Retrieve a single household by ID.",
	}, getByIDHandler("/reporting/households"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_households",
		Description: "Filter households by criteria such as name, firm_id.",
	}, filterHandler("/reporting/households/filter"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_household",
		Description: "Create a new household.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateInput) (*mcp.CallToolResult, any, error) {
		data, err := apiClient.Request("POST", "/reporting/households", input.Data, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remap_households",
		Description: "Assign accounts to households. Allows bulk reassignment of accounts across households.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input HouseholdRemapInput) (*mcp.CallToolResult, any, error) {
		data, err := apiClient.Request("POST", "/reporting/households/remap", input.Assignments, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	// ==================== BENCHMARKS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_benchmarks",
		Description: "List all benchmarks used for performance comparison.",
	}, listHandler("/reporting/benchmarks"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_benchmark",
		Description: "Retrieve a single benchmark by ID.",
	}, getByIDHandler("/reporting/benchmarks"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_benchmarks",
		Description: "Filter benchmarks by criteria.",
	}, filterHandler("/reporting/benchmarks/filter"))

	// ==================== ASSET CLASSIFICATIONS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_asset_classifications",
		Description: "List all asset classification overlays used to categorize securities.",
	}, listHandler("/reporting/asset-classifications"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_asset_classifications",
		Description: "Filter asset classifications by criteria.",
	}, filterHandler("/reporting/asset-classifications/filter"))

	// ==================== TARGET ALLOCATIONS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_target_allocations",
		Description: "List all target allocation models that define desired portfolio composition.",
	}, listHandler("/reporting/target-allocations"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_target_allocations",
		Description: "Filter target allocations by criteria.",
	}, filterHandler("/reporting/target-allocations/filter"))

	// ==================== INVESTMENT MODELS & STRATEGIES ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_models",
		Description: "List all investment models.",
	}, listHandler("/investment-management/models"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_model",
		Description: "Retrieve a single investment model by ID.",
	}, getByIDHandler("/investment-management/models"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_strategies",
		Description: "List all investment strategies.",
	}, listHandler("/investment-management/strategies"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_strategy",
		Description: "Retrieve a single investment strategy by ID.",
	}, getByIDHandler("/investment-management/strategies"))

	// ==================== BILLING ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_groups",
		Description: "List all billing groups that organize accounts for fee calculation.",
	}, listHandler("/billing/groups"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_fee_structures",
		Description: "List all fee structures defining how advisory fees are calculated.",
	}, listHandler("/billing/fee-structures"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_invoices",
		Description: "List all generated billing invoices.",
	}, listHandler("/billing/invoices"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_invoices",
		Description: "Filter invoices by criteria such as billing_group_id, period, status.",
	}, filterHandler("/billing/invoices/filter"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_reports",
		Description: "List all billing reports (billing job results).",
	}, listHandler("/billing/reports"))

	// ==================== PRINTABLE REPORTS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_printable_reports",
		Description: "List all generated printable PDF reports.",
	}, listHandler("/reporting/printable-reports"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_printable_reports",
		Description: "Filter printable reports by criteria.",
	}, filterHandler("/reporting/printable-reports/filter"))

	// ==================== WEB REPORTS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_web_report",
		Description: "Run a web report with specified parameters. Returns report data for display.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input WebReportInput) (*mcp.CallToolResult, any, error) {
		body := map[string]interface{}{
			"report_type": input.ReportType,
		}
		for k, v := range input.Params {
			body[k] = v
		}
		data, err := apiClient.Request("POST", "/reporting/web-reports", body, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	// ==================== JOBS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_jobs",
		Description: "List all background jobs (billing runs, report generation, etc.).",
	}, listHandler("/jobs/"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_job",
		Description: "Retrieve a single job by ID to check its status and results.",
	}, getByIDHandler("/jobs"))

	// ==================== SOURCE DATA STATUS ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "account_summary_status",
		Description: "Get account meta status summary showing data feed health across custodians.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		data, err := apiClient.Request("GET", "/status/account-summary", nil, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "source_data_status",
		Description: "Check the status of source data feeds and processing pipelines.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input FilterInput) (*mcp.CallToolResult, any, error) {
		data, err := apiClient.Request("POST", "/status/source-data", input.Filter, nil)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})

	// ==================== IDC INDEXES ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_indexes",
		Description: "List all IDC market indexes available for benchmarking.",
	}, listHandler("/data/idc/indexes"))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "filter_indexes",
		Description: "Filter IDC indexes by criteria.",
	}, filterHandler("/data/idc/indexes/filter"))

	// ==================== ADVISOR CODES ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_advisor_codes",
		Description: "List all advisor/rep codes associated with your firm.",
	}, listHandler("/org/advisor-codes"))

	// ==================== GENERIC API REQUEST ====================
	mcp.AddTool(server, &mcp.Tool{
		Name:        "api_request",
		Description: "Make a raw API request to any BridgeFT endpoint. Use this for endpoints not covered by other tools, or for PUT/DELETE operations. Base URL is https://api.bridgeft.com/v2. Example paths: /account-management/accounts, /billing/groups, /reporting/households.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input APIRequestInput) (*mcp.CallToolResult, any, error) {
		var body interface{}
		if input.Body != nil {
			body = input.Body
		}
		data, err := apiClient.Request(input.Method, input.Path, body, input.Query)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(data), nil, nil
	})
}
