package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// TopupPlan describes a top-up plan or amount tier.
type TopupPlan struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	BonusAmount float64 `json:"bonus_amount,omitempty"`
	Description string  `json:"description,omitempty"`
}

func (p *TopupPlan) UnmarshalJSON(data []byte) error {
	type alias TopupPlan
	var raw struct {
		alias
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = TopupPlan(raw.alias)
	p.ID = stringFromJSON(raw.ID)
	return nil
}

// Topup describes a top-up transaction.
type Topup struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at,omitempty"`
	OrderID   string  `json:"order_id,omitempty"`
}

func (t *Topup) UnmarshalJSON(data []byte) error {
	type alias Topup
	var raw struct {
		alias
		ID      json.RawMessage `json:"id"`
		OrderID json.RawMessage `json:"order_id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*t = Topup(raw.alias)
	t.ID = stringFromJSON(raw.ID)
	if t.OrderID == "" {
		t.OrderID = stringFromJSON(raw.OrderID)
	}
	return nil
}

// CreateTopupCheckoutRequest creates a top-up checkout session.
type CreateTopupCheckoutRequest struct {
	Amount      float64 `json:"amount,omitempty"`
	TopupPlanID string  `json:"topup_plan_id,omitempty"`
	SuccessURL  string  `json:"success_url"`
	CancelURL   string  `json:"cancel_url"`
}

// CreateTopupCheckoutResponse contains checkout session details.
type CreateTopupCheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
	OrderID     string `json:"order_id,omitempty"`
}

// CreateTopupInAppRequest creates an in-app top-up.
type CreateTopupInAppRequest struct {
	Amount      float64 `json:"amount,omitempty"`
	TopupPlanID string  `json:"topup_plan_id,omitempty"`
}

// CreateTopupInAppResponse contains payment intent details.
type CreateTopupInAppResponse struct {
	PaymentIntentID string `json:"payment_intent_id"`
	ClientSecret    string `json:"client_secret"`
	RequiresAction  bool   `json:"requires_action"`
	Status          string `json:"status"`
}

// ConfirmTopupInAppRequest confirms in-app top-up payment.
type ConfirmTopupInAppRequest struct {
	PaymentIntentID string `json:"payment_intent_id"`
}

// ListTopupPlans lists available top-up plans.
func (c *Client) ListTopupPlans(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[TopupPlan], *owlvigil.ResponseMeta, error) {
	var out ListResponse[TopupPlan]
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/topup-plans", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateTopupCheckout creates a top-up checkout session.
func (c *Client) CreateTopupCheckout(ctx context.Context, req *CreateTopupCheckoutRequest, reqOpts ...owlvigil.RequestOption) (*CreateTopupCheckoutResponse, *owlvigil.ResponseMeta, error) {
	var out CreateTopupCheckoutResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/topups/checkout", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateTopupInApp creates an in-app top-up payment.
func (c *Client) CreateTopupInApp(ctx context.Context, req *CreateTopupInAppRequest, reqOpts ...owlvigil.RequestOption) (*CreateTopupInAppResponse, *owlvigil.ResponseMeta, error) {
	var out CreateTopupInAppResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/topups/in-app", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ConfirmTopupInApp confirms in-app top-up payment after 3DS.
func (c *Client) ConfirmTopupInApp(ctx context.Context, req *ConfirmTopupInAppRequest, reqOpts ...owlvigil.RequestOption) (*Topup, *owlvigil.ResponseMeta, error) {
	var out Topup
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/topups/in-app/confirm", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListTopups lists top-up transactions.
func (c *Client) ListTopups(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Topup], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Topup]
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/topups", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetTopup retrieves top-up details by ID.
func (c *Client) GetTopup(ctx context.Context, topupID string, reqOpts ...owlvigil.RequestOption) (*Topup, *owlvigil.ResponseMeta, error) {
	var out Topup
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/topups/"+url.PathEscape(topupID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
