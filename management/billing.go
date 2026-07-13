package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// BillingOverview contains billing overview information.
type BillingOverview struct {
	WorkspaceID          int64          `json:"workspace_id,omitempty"`
	Balance              *Balance       `json:"balance,omitempty"`
	BalanceAmount        float64        `json:"balance_amount,omitempty"`
	CurrentMonthCost     float64        `json:"current_month_cost,omitempty"`
	CurrentMonthTokens   int64          `json:"current_month_tokens,omitempty"`
	CurrentMonthRequests int64          `json:"current_month_requests,omitempty"`
	Subscription         *Subscription  `json:"subscription,omitempty"`
	DefaultPaymentMethod *PaymentMethod `json:"default_payment_method,omitempty"`
	RecentInvoices       []Invoice      `json:"recent_invoices,omitempty"`
}

func (b *BillingOverview) UnmarshalJSON(data []byte) error {
	type alias BillingOverview
	var raw struct {
		alias
		Balance json.RawMessage `json:"balance"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*b = BillingOverview(raw.alias)
	if len(raw.Balance) == 0 || string(raw.Balance) == "null" {
		return nil
	}
	var balance Balance
	if err := json.Unmarshal(raw.Balance, &balance); err == nil {
		b.Balance = &balance
		b.BalanceAmount = balance.Amount
		return nil
	}
	var amount float64
	if err := json.Unmarshal(raw.Balance, &amount); err == nil {
		b.BalanceAmount = amount
		b.Balance = &Balance{Amount: amount}
	}
	return nil
}

// BillingDetails contains billing contact and address information.
type BillingDetails struct {
	CompanyName  string   `json:"company_name,omitempty"`
	Email        string   `json:"email,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	Address      *Address `json:"address,omitempty"`
	CCRecipients []string `json:"cc_recipients,omitempty"`
}

// Address describes a billing address.
type Address struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

// UpdateBillingDetailsRequest updates billing details.
type UpdateBillingDetailsRequest struct {
	CompanyName  *string  `json:"company_name,omitempty"`
	Email        *string  `json:"email,omitempty"`
	Phone        *string  `json:"phone,omitempty"`
	Address      *Address `json:"address,omitempty"`
	CCRecipients []string `json:"cc_recipients,omitempty"`
}

// GetBillingOverview retrieves billing overview.
func (c *Client) GetBillingOverview(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*BillingOverview, *owlvigil.ResponseMeta, error) {
	var out BillingOverview
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/overview", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetBillingDetails retrieves billing details for a workspace.
func (c *Client) GetBillingDetails(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*BillingDetails, *owlvigil.ResponseMeta, error) {
	var out BillingDetails
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/billing-details"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateBillingDetails updates billing details for a workspace.
func (c *Client) UpdateBillingDetails(ctx context.Context, workspaceID int64, req *UpdateBillingDetailsRequest, reqOpts ...owlvigil.RequestOption) (*BillingDetails, *owlvigil.ResponseMeta, error) {
	var out BillingDetails
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/billing-details"
	meta, err := c.http.Do(ctx, http.MethodPut, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetInvoice retrieves invoice details by ID.
func (c *Client) GetInvoice(ctx context.Context, invoiceID string, reqOpts ...owlvigil.RequestOption) (*Invoice, *owlvigil.ResponseMeta, error) {
	var out Invoice
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/invoices/"+url.PathEscape(invoiceID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
