package management

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	owlvigil "github.com/owlvigil/owlvigil-go"
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
	Plan         any         `json:"plan"`
	Items        []QuotaItem `json:"items,omitempty"`
	ResetTime    string      `json:"reset_time,omitempty"`
	RequestQuota *Quota      `json:"request_quota,omitempty"`
	TokenQuota   *Quota      `json:"token_quota,omitempty"`
	CostQuota    *Quota      `json:"cost_quota,omitempty"`
}

type QuotaItem struct {
	Key   string  `json:"key"`
	Label string  `json:"label,omitempty"`
	Used  float64 `json:"used"`
	Limit float64 `json:"limit"`
}

// QuotaUsage describes quota usage by scope.
type QuotaUsage struct {
	Teams       any `json:"teams,omitempty"`
	Members     any `json:"members,omitempty"`
	GatewayKeys any `json:"gateway_keys,omitempty"`
	APIKeys     any `json:"api_keys,omitempty"`
}

// Balance describes the current billing balance.
type Balance struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Invoice describes an invoice returned by billing APIs.
type Invoice struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	DueDate     string  `json:"due_date,omitempty"`
	PaidAt      string  `json:"paid_at,omitempty"`
	InvoiceURL  string  `json:"invoice_url,omitempty"`
	DownloadURL string  `json:"download_url,omitempty"`
}

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
		i.Amount = *raw.AmountDue
	}
	if i.InvoiceURL == "" {
		i.InvoiceURL = raw.HostedURL
	}
	if i.DownloadURL == "" {
		i.DownloadURL = raw.PDFURL
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

// ListInvoices lists billing invoices.
func (c *Client) ListInvoices(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Invoice], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Invoice]
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/invoices", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
