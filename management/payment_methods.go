package management

import (
	"context"
	"net/http"
	"net/url"

	owlvigil "github.com/owlvigil/owlvigil-go"
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
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/payment-methods/setup-intent", nil, nil, &out, reqOpts...)
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
