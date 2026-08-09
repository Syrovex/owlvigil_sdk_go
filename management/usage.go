package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// UsageSummary contains aggregated Gateway usage.
type UsageSummary struct {
	Requests int64   `json:"requests"`
	Tokens   int64   `json:"tokens"`
	Cost     float64 `json:"cost"`
}

// UsageRecord describes a detailed usage record.
type UsageRecord struct {
	ID        string  `json:"id"`
	Timestamp string  `json:"timestamp"`
	Model     string  `json:"model"`
	Provider  string  `json:"provider,omitempty"`
	Requests  int64   `json:"requests"`
	Tokens    int64   `json:"tokens"`
	Cost      float64 `json:"cost"`
}

// Quota describes current quota consumption.
type Quota struct {
	Limit     float64 `json:"limit"`
	Used      float64 `json:"used"`
	Remaining float64 `json:"remaining"`
}

// QuotaSummary describes quota summary for a workspace.
type QuotaSummary struct {
	WorkspaceID  int64       `json:"workspace_id,omitempty"`
	Plan         QuotaPlan   `json:"plan"`
	Items        []QuotaItem `json:"items,omitempty"`
	ResetTime    string      `json:"reset_time,omitempty"`
	RequestQuota *Quota      `json:"request_quota,omitempty"`
	TokenQuota   *Quota      `json:"token_quota,omitempty"`
	CostQuota    *Quota      `json:"cost_quota,omitempty"`
}

// QuotaPlan identifies the workspace plan and its machine-readable limits.
type QuotaPlan struct {
	Name   string           `json:"name"`
	Slug   string           `json:"slug"`
	Quotas map[string]int64 `json:"quotas"`
}

// QuotaItem describes one aggregate quota status.
type QuotaItem struct {
	Key   string `json:"key"`
	Label string `json:"label,omitempty"`
	Used  int64  `json:"used"`
	Limit int64  `json:"limit"`
	Unit  string `json:"unit,omitempty"`
}

// QuotaCounter contains one resource's used and allowed quantities.
type QuotaCounter struct {
	Used  int   `json:"used"`
	Limit int64 `json:"limit"`
}

// QuotaUsage describes quota usage by scope.
type QuotaUsage struct {
	Teams   QuotaCounter `json:"teams"`
	Members QuotaCounter `json:"members"`
	APIKeys QuotaCounter `json:"api_keys"`

	// GatewayKeys is the pre-refactor alias for APIKeys.
	GatewayKeys QuotaCounter `json:"-"`
}

// UnmarshalJSON synchronizes the current api_keys field with the legacy
// GatewayKeys alias.
func (q *QuotaUsage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Teams       QuotaCounter `json:"teams"`
		Members     QuotaCounter `json:"members"`
		APIKeys     QuotaCounter `json:"api_keys"`
		GatewayKeys QuotaCounter `json:"gateway_keys"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	q.Teams = raw.Teams
	q.Members = raw.Members
	q.APIKeys = raw.APIKeys
	if q.APIKeys == (QuotaCounter{}) {
		q.APIKeys = raw.GatewayKeys
	}
	q.GatewayKeys = q.APIKeys
	return nil
}

// Balance describes the current billing balance.
type Balance struct {
	WorkspaceID int64   `json:"workspace_id,omitempty"`
	Balance     float64 `json:"balance"`

	// Amount and Currency are retained for pre-refactor responses.
	Amount   float64 `json:"amount,omitempty"`
	Currency string  `json:"currency,omitempty"`
}

// UnmarshalJSON accepts both the current balance response and the legacy
// amount/currency response.
func (b *Balance) UnmarshalJSON(data []byte) error {
	var raw struct {
		WorkspaceID int64    `json:"workspace_id,omitempty"`
		Balance     *float64 `json:"balance,omitempty"`
		Amount      *float64 `json:"amount,omitempty"`
		Currency    string   `json:"currency,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	b.WorkspaceID = raw.WorkspaceID
	b.Currency = raw.Currency
	if raw.Balance != nil {
		b.Balance = *raw.Balance
		b.Amount = *raw.Balance
		return nil
	}
	if raw.Amount != nil {
		b.Amount = *raw.Amount
		b.Balance = *raw.Amount
	}
	return nil
}

// Invoice describes an invoice returned by billing APIs.
type Invoice struct {
	ID               string  `json:"id"`
	WorkspaceID      int64   `json:"workspace_id,omitempty"`
	StripeInvoiceID  string  `json:"stripe_invoice_id,omitempty"`
	InvoiceNumber    *string `json:"invoice_number,omitempty"`
	AmountDue        float64 `json:"amount_due,omitempty"`
	AmountPaid       float64 `json:"amount_paid,omitempty"`
	Currency         string  `json:"currency"`
	Status           string  `json:"status"`
	InvoicePDFURL    string  `json:"invoice_pdf_url,omitempty"`
	HostedInvoiceURL string  `json:"hosted_invoice_url,omitempty"`
	PeriodStart      string  `json:"period_start,omitempty"`
	PeriodEnd        string  `json:"period_end,omitempty"`
	DueDate          string  `json:"due_date,omitempty"`
	PaidAt           string  `json:"paid_at,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`

	// Legacy aliases retained for source compatibility.
	Amount      float64 `json:"amount,omitempty"`
	InvoiceURL  string  `json:"invoice_url,omitempty"`
	DownloadURL string  `json:"download_url,omitempty"`
}

// UnmarshalJSON accepts numeric and string invoice IDs and populates legacy aliases.
func (i *Invoice) UnmarshalJSON(data []byte) error {
	type alias Invoice
	var raw struct {
		alias
		ID        json.RawMessage `json:"id"`
		AmountDue *float64        `json:"amount_due,omitempty"`
		PDFURL    string          `json:"invoice_pdf_url,omitempty"`
		HostedURL string          `json:"hosted_invoice_url,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*i = Invoice(raw.alias)
	i.ID = stringFromJSON(raw.ID)
	if raw.AmountDue != nil {
		i.AmountDue = *raw.AmountDue
		i.Amount = *raw.AmountDue
	}
	if i.InvoicePDFURL == "" {
		i.InvoicePDFURL = raw.PDFURL
	}
	if i.HostedInvoiceURL == "" {
		i.HostedInvoiceURL = raw.HostedURL
	}
	if i.InvoiceURL == "" {
		i.InvoiceURL = i.HostedInvoiceURL
		if i.InvoiceURL == "" {
			i.InvoiceURL = raw.HostedURL
		}
	}
	if i.DownloadURL == "" {
		i.DownloadURL = i.InvoicePDFURL
		if i.DownloadURL == "" {
			i.DownloadURL = raw.PDFURL
		}
	}
	return nil
}

// ListUsage lists detailed usage records.
func (c *Client) ListUsage(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[UsageRecord], *owlvigil.ResponseMeta, error) {
	var out ListResponse[UsageRecord]
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/usage", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetUsageSummary retrieves aggregated Gateway usage.
func (c *Client) GetUsageSummary(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*UsageSummary, *owlvigil.ResponseMeta, error) {
	var out UsageSummary
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/usage/summary", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetQuota retrieves quota information.
func (c *Client) GetQuota(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*Quota, *owlvigil.ResponseMeta, error) {
	var out Quota
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/quota", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetQuotaSummary retrieves workspace quota summary.
func (c *Client) GetQuotaSummary(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*QuotaSummary, *owlvigil.ResponseMeta, error) {
	var out QuotaSummary
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/quota-summary"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetQuotaUsage retrieves workspace quota usage breakdown.
func (c *Client) GetQuotaUsage(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*QuotaUsage, *owlvigil.ResponseMeta, error) {
	var out QuotaUsage
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/quota-usage"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetBalance retrieves the billing balance.
func (c *Client) GetBalance(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*Balance, *owlvigil.ResponseMeta, error) {
	var out Balance
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/balance", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetBalanceForWorkspace retrieves the billing balance for a workspace.
func (c *Client) GetBalanceForWorkspace(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*Balance, *owlvigil.ResponseMeta, error) {
	var out Balance
	query := url.Values{"workspace_id": {strconv.FormatInt(workspaceID, 10)}}
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/balance", query, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListInvoices lists billing invoices.
func (c *Client) ListInvoices(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Invoice], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Invoice]
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/invoices", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListInvoicesForWorkspace lists billing invoices for a workspace.
func (c *Client) ListInvoicesForWorkspace(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Invoice], *owlvigil.ResponseMeta, error) {
	query := opts.values()
	query.Set("workspace_id", strconv.FormatInt(workspaceID, 10))
	var out ListResponse[Invoice]
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/invoices", query, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
