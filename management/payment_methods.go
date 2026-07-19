package management

import (
	"context"
	"net/http"
	"net/url"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// PaymentMethod describes a saved payment method.
type PaymentMethod struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Brand     string                 `json:"brand,omitempty"`
	Last4     string                 `json:"last4,omitempty"`
	ExpMonth  int                    `json:"exp_month,omitempty"`
	ExpYear   int                    `json:"exp_year,omitempty"`
	IsDefault bool                   `json:"is_default"`
	CreatedAt string                 `json:"created_at,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
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
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/payment-methods", opts.values(), nil, &out, reqOpts...)
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
	return c.http.Do(ctx, http.MethodDelete, "/billing/payment-methods/"+url.PathEscape(paymentMethodID), nil, nil, nil, reqOpts...)
}
