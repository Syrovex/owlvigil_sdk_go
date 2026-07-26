package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// PaymentMethod describes a saved payment method.
type PaymentMethod struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Brand        string         `json:"brand,omitempty"`
	Last4        string         `json:"last4,omitempty"`
	ExpiryMonth  int            `json:"expiry_month,omitempty"`
	ExpiryYear   int            `json:"expiry_year,omitempty"`
	IsDefault    bool           `json:"is_default"`
	CreatedAt    string         `json:"created_at,omitempty"`
	BillingEmail string         `json:"billing_email,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`

	// ExpMonth and ExpYear are pre-refactor aliases.
	ExpMonth int `json:"-"`
	ExpYear  int `json:"-"`
}

// UnmarshalJSON decodes the current expiry field names and synchronizes the
// pre-refactor aliases.
func (p *PaymentMethod) UnmarshalJSON(data []byte) error {
	type alias PaymentMethod
	var raw struct {
		alias
		LegacyExpMonth int `json:"exp_month,omitempty"`
		LegacyExpYear  int `json:"exp_year,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = PaymentMethod(raw.alias)
	if p.ExpiryMonth == 0 {
		p.ExpiryMonth = raw.LegacyExpMonth
	}
	if p.ExpiryYear == 0 {
		p.ExpiryYear = raw.LegacyExpYear
	}
	p.ExpMonth = p.ExpiryMonth
	p.ExpYear = p.ExpiryYear
	return nil
}

// SetupIntent describes a Stripe SetupIntent for adding payment methods.
type SetupIntent struct {
	SetupIntentID string `json:"setup_intent_id"`
	ClientSecret  string `json:"client_secret"`
	Status        string `json:"status"`
}

// CreatePaymentMethodSetupIntentRequest creates a SetupIntent for a workspace.
type CreatePaymentMethodSetupIntentRequest struct {
	WorkspaceID int64 `json:"workspace_id"`
}

// SavePaymentMethodRequest saves a payment method.
type SavePaymentMethodRequest struct {
	PaymentMethodID string `json:"payment_method_id"`
	SetAsDefault    bool   `json:"set_as_default,omitempty"`
}

// ListPaymentMethods lists saved payment methods.
func (c *Client) ListPaymentMethods(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[PaymentMethod], *owlvigil.ResponseMeta, error) {
	var out ListResponse[PaymentMethod]
	// Only workspace_id is declared; callers may supply it as a request option.
	_ = opts
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/payment-methods", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListPaymentMethodsForWorkspace lists saved payment methods for a workspace.
func (c *Client) ListPaymentMethodsForWorkspace(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*ListResponse[PaymentMethod], *owlvigil.ResponseMeta, error) {
	var out ListResponse[PaymentMethod]
	query := url.Values{"workspace_id": {strconv.FormatInt(workspaceID, 10)}}
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/payment-methods", query, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreatePaymentMethodSetupIntent creates a Stripe SetupIntent.
func (c *Client) CreatePaymentMethodSetupIntent(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*SetupIntent, *owlvigil.ResponseMeta, error) {
	var out SetupIntent
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/payment-methods/setup-intent", nil, struct{}{}, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreatePaymentMethodSetupIntentForWorkspace creates a Stripe SetupIntent for
// the specified workspace. Use the returned client secret with Stripe.js to
// collect and confirm the card; do not send card data through the SDK.
func (c *Client) CreatePaymentMethodSetupIntentForWorkspace(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*SetupIntent, *owlvigil.ResponseMeta, error) {
	var out SetupIntent
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/payment-methods/setup-intent", nil, CreatePaymentMethodSetupIntentRequest{WorkspaceID: workspaceID}, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// SavePaymentMethod saves a confirmed payment method.
func (c *Client) SavePaymentMethod(ctx context.Context, req *SavePaymentMethodRequest, reqOpts ...owlvigil.RequestOption) (*PaymentMethod, *owlvigil.ResponseMeta, error) {
	var out PaymentMethod
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/payment-methods", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// SetDefaultPaymentMethod sets a payment method as default.
func (c *Client) SetDefaultPaymentMethod(ctx context.Context, paymentMethodID string, reqOpts ...owlvigil.RequestOption) (*PaymentMethod, *owlvigil.ResponseMeta, error) {
	var out PaymentMethod
	path := "/billing/payment-methods/" + url.PathEscape(paymentMethodID) + "/default"
	meta, err := c.http.Do(ctx, http.MethodPut, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeletePaymentMethod removes a saved payment method.
func (c *Client) DeletePaymentMethod(ctx context.Context, paymentMethodID string, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.DeletePaymentMethodWithResult(ctx, paymentMethodID, reqOpts...)
	return meta, err
}

// DeletePaymentMethodWithResult removes a payment method and returns confirmation.
func (c *Client) DeletePaymentMethodWithResult(ctx context.Context, paymentMethodID string, reqOpts ...owlvigil.RequestOption) (*DeleteResponse, *owlvigil.ResponseMeta, error) {
	var out DeleteResponse
	meta, err := c.http.Do(ctx, http.MethodDelete, "/billing/payment-methods/"+url.PathEscape(paymentMethodID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
