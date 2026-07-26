package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// BillingOverview contains billing overview information.
type BillingOverview struct {
	WorkspaceID            int64          `json:"workspace_id,omitempty"`
	Balance                *Balance       `json:"balance,omitempty"`
	BalanceAmount          float64        `json:"balance_amount,omitempty"`
	MonthlyBudgetLimit     *float64       `json:"monthly_budget_limit,omitempty"`
	CurrentMonthCost       float64        `json:"current_month_cost,omitempty"`
	CurrentMonthTokens     int64          `json:"current_month_tokens,omitempty"`
	CurrentMonthRequests   int64          `json:"current_month_requests,omitempty"`
	BillingDetails         map[string]any `json:"billing_details,omitempty"`
	RecentInvoices         []Invoice      `json:"recent_invoices,omitempty"`
	RecentInvoicesPageInfo PageInfo       `json:"recent_invoices_page_info"`

	// Legacy response fields retained for source compatibility.
	Subscription         *Subscription  `json:"subscription,omitempty"`
	DefaultPaymentMethod *PaymentMethod `json:"default_payment_method,omitempty"`
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
		b.Balance = &Balance{
			WorkspaceID: b.WorkspaceID,
			Balance:     amount,
			Amount:      amount,
		}
	}
	return nil
}

// BillingDetails contains billing contact and address information.
type BillingDetails struct {
	WorkspaceID int64          `json:"workspace_id,omitempty"`
	Details     map[string]any `json:"details,omitempty"`

	Name         string   `json:"name,omitempty"`
	Email        string   `json:"email,omitempty"`
	TaxID        string   `json:"tax_id,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	AddressText  string   `json:"address,omitempty"`
	CCRecipients []string `json:"cc_recipients,omitempty"`

	// CompanyName and Address retain the pre-refactor SDK response aliases.
	CompanyName string   `json:"-"`
	Address     *Address `json:"-"`
}

// UnmarshalJSON accepts both the current {"workspace_id":...,"details":{...}}
// response and the pre-refactor flattened billing-details response.
func (d *BillingDetails) UnmarshalJSON(data []byte) error {
	type wireDetails struct {
		WorkspaceID  int64           `json:"workspace_id,omitempty"`
		Details      json.RawMessage `json:"details,omitempty"`
		Name         string          `json:"name,omitempty"`
		CompanyName  string          `json:"company_name,omitempty"`
		Email        string          `json:"email,omitempty"`
		TaxID        string          `json:"tax_id,omitempty"`
		Phone        string          `json:"phone,omitempty"`
		Address      json.RawMessage `json:"address,omitempty"`
		CCRecipients []string        `json:"cc_recipients,omitempty"`
	}
	var raw wireDetails
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*d = BillingDetails{WorkspaceID: raw.WorkspaceID}
	apply := func(value wireDetails) error {
		d.Name = value.Name
		d.CompanyName = value.CompanyName
		if d.Name == "" {
			d.Name = d.CompanyName
		}
		if d.CompanyName == "" {
			d.CompanyName = d.Name
		}
		d.Email = value.Email
		d.TaxID = value.TaxID
		d.Phone = value.Phone
		d.CCRecipients = value.CCRecipients
		if len(value.Address) == 0 || string(value.Address) == "null" {
			return nil
		}
		if err := json.Unmarshal(value.Address, &d.AddressText); err == nil {
			return nil
		}
		var address Address
		if err := json.Unmarshal(value.Address, &address); err != nil {
			return err
		}
		d.Address = &address
		return nil
	}

	if len(raw.Details) > 0 && string(raw.Details) != "null" {
		if err := json.Unmarshal(raw.Details, &d.Details); err != nil {
			return err
		}
		var nested wireDetails
		if err := json.Unmarshal(raw.Details, &nested); err != nil {
			return err
		}
		return apply(nested)
	}
	return apply(raw)
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
	BillingDetails *BillingContact `json:"billing_details,omitempty"`
	Name           string          `json:"name,omitempty"`
	Email          string          `json:"email,omitempty"`
	TaxID          string          `json:"tax_id,omitempty"`
	Phone          string          `json:"phone,omitempty"`
	AddressText    string          `json:"address,omitempty"`
	CCRecipients   []string        `json:"cc_recipients,omitempty"`

	// CompanyName and Address are legacy aliases retained for source
	// compatibility. MarshalJSON maps them to the current API fields.
	CompanyName *string  `json:"-"`
	Address     *Address `json:"-"`
}

// BillingContact contains the current Open API billing contact shape.
type BillingContact struct {
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	TaxID        string   `json:"tax_id"`
	Phone        string   `json:"phone"`
	Address      string   `json:"address"`
	CCRecipients []string `json:"cc_recipients"`
}

// MarshalJSON emits the current Open API billing-details request while
// accepting the pre-refactor CompanyName and structured Address aliases.
func (r UpdateBillingDetailsRequest) MarshalJSON() ([]byte, error) {
	name := r.Name
	if name == "" && r.CompanyName != nil {
		name = *r.CompanyName
	}
	address := r.AddressText
	if address == "" && r.Address != nil {
		encoded, err := json.Marshal(r.Address)
		if err != nil {
			return nil, err
		}
		address = string(encoded)
	}
	type wireRequest struct {
		BillingDetails *BillingContact `json:"billing_details,omitempty"`
		Name           string          `json:"name,omitempty"`
		Email          string          `json:"email,omitempty"`
		TaxID          string          `json:"tax_id,omitempty"`
		Phone          string          `json:"phone,omitempty"`
		Address        string          `json:"address,omitempty"`
		CCRecipients   []string        `json:"cc_recipients,omitempty"`
	}
	return json.Marshal(wireRequest{
		BillingDetails: r.BillingDetails,
		Name:           name,
		Email:          r.Email,
		TaxID:          r.TaxID,
		Phone:          r.Phone,
		Address:        address,
		CCRecipients:   r.CCRecipients,
	})
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

// GetBillingOverviewForWorkspace retrieves billing overview for a workspace.
func (c *Client) GetBillingOverviewForWorkspace(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*BillingOverview, *owlvigil.ResponseMeta, error) {
	var out BillingOverview
	query := url.Values{"workspace_id": {strconv.FormatInt(workspaceID, 10)}}
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/overview", query, nil, &out, reqOpts...)
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

// GetInvoiceForWorkspace retrieves invoice details by ID for a workspace.
func (c *Client) GetInvoiceForWorkspace(ctx context.Context, workspaceID int64, invoiceID string, reqOpts ...owlvigil.RequestOption) (*Invoice, *owlvigil.ResponseMeta, error) {
	var out Invoice
	query := url.Values{"workspace_id": {strconv.FormatInt(workspaceID, 10)}}
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/invoices/"+url.PathEscape(invoiceID), query, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
