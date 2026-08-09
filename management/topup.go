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
	ID              string   `json:"id"`
	Name            string   `json:"name,omitempty"`
	Currency        string   `json:"currency"`
	MinAmount       float64  `json:"min_amount,omitempty"`
	MaxAmount       float64  `json:"max_amount,omitempty"`
	FeeRate         float64  `json:"fee_rate,omitempty"`
	PaymentChannels []string `json:"payment_channels,omitempty"`
	CustomAmount    bool     `json:"custom_amount,omitempty"`

	// Legacy fields retained for source compatibility.
	Amount      float64 `json:"amount,omitempty"`
	BonusAmount float64 `json:"bonus_amount,omitempty"`
	Description string  `json:"description,omitempty"`
}

// UnmarshalJSON accepts numeric and string top-up plan IDs.
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
	ID                    string         `json:"id"`
	WorkspaceID           *int64         `json:"workspace_id,omitempty"`
	UserID                int            `json:"user_id,omitempty"`
	Amount                float64        `json:"amount"`
	PayAmount             float64        `json:"pay_amount,omitempty"`
	Currency              string         `json:"currency"`
	OrderType             string         `json:"order_type,omitempty"`
	PlanID                *int64         `json:"plan_id,omitempty"`
	Status                string         `json:"status"`
	OutTradeNo            string         `json:"out_trade_no,omitempty"`
	PaymentType           string         `json:"payment_type,omitempty"`
	PaymentTradeNo        string         `json:"payment_trade_no,omitempty"`
	PayURL                *string        `json:"pay_url,omitempty"`
	RechargeCode          string         `json:"recharge_code,omitempty"`
	ProviderSnapshot      map[string]any `json:"provider_snapshot,omitempty"`
	StripeSessionID       string         `json:"stripe_session_id,omitempty"`
	StripePaymentIntentID string         `json:"stripe_payment_intent_id,omitempty"`
	ExpiresAt             string         `json:"expires_at,omitempty"`
	PaidAt                string         `json:"paid_at,omitempty"`
	CompletedAt           string         `json:"completed_at,omitempty"`
	FailedAt              string         `json:"failed_at,omitempty"`
	FailedReason          *string        `json:"failed_reason,omitempty"`
	CreatedAt             string         `json:"created_at,omitempty"`
	UpdatedAt             string         `json:"updated_at,omitempty"`

	// OrderID is the legacy alias for ID.
	OrderID string `json:"order_id,omitempty"`
}

// UnmarshalJSON accepts numeric and string top-up IDs and populates OrderID.
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
	if t.OrderID == "" {
		t.OrderID = t.ID
	}
	return nil
}

// CreateTopupCheckoutRequest creates a top-up checkout session.
type CreateTopupCheckoutRequest struct {
	WorkspaceID int64   `json:"workspace_id"`
	Amount      float64 `json:"amount"`
	TopupPlanID string  `json:"topup_plan_id,omitempty"`
	SuccessURL  string  `json:"success_url"`
	CancelURL   string  `json:"cancel_url"`
}

// CreateTopupCheckoutResponse contains checkout session details.
type CreateTopupCheckoutResponse struct {
	OrderID      string  `json:"order_id,omitempty"`
	WorkspaceID  int64   `json:"workspace_id,omitempty"`
	OutTradeNo   string  `json:"out_trade_no,omitempty"`
	CheckoutURL  string  `json:"checkout_url"`
	SessionID    string  `json:"session_id"`
	Amount       float64 `json:"amount,omitempty"`
	Currency     string  `json:"currency,omitempty"`
	OrderType    string  `json:"order_type,omitempty"`
	Status       string  `json:"status,omitempty"`
	ExpiresAt    string  `json:"expires_at,omitempty"`
	StripeMode   string  `json:"stripe_mode,omitempty"`
	Subscription string  `json:"subscription,omitempty"`
}

// UnmarshalJSON accepts numeric and string order IDs.
func (r *CreateTopupCheckoutResponse) UnmarshalJSON(data []byte) error {
	type alias CreateTopupCheckoutResponse
	var raw struct {
		alias
		OrderID json.RawMessage `json:"order_id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = CreateTopupCheckoutResponse(raw.alias)
	r.OrderID = stringFromJSON(raw.OrderID)
	return nil
}

// CreateTopupInAppRequest creates an in-app top-up.
type CreateTopupInAppRequest struct {
	WorkspaceID int64   `json:"workspace_id"`
	Amount      float64 `json:"amount"`
	TopupPlanID string  `json:"topup_plan_id,omitempty"`
	ReturnURL   string  `json:"return_url"`
}

// CreateTopupInAppResponse contains payment intent details.
type CreateTopupInAppResponse struct {
	OrderID         string  `json:"order_id,omitempty"`
	WorkspaceID     int64   `json:"workspace_id,omitempty"`
	PaymentIntentID string  `json:"payment_intent_id"`
	ClientSecret    string  `json:"client_secret"`
	RequiresAction  bool    `json:"requires_action"`
	Status          string  `json:"status"`
	OrderType       string  `json:"order_type,omitempty"`
	Amount          float64 `json:"amount,omitempty"`
	Currency        string  `json:"currency,omitempty"`
	Balance         float64 `json:"balance,omitempty"`
	TotalRecharged  float64 `json:"total_recharged,omitempty"`
	Message         string  `json:"message,omitempty"`
}

// UnmarshalJSON accepts numeric and string order IDs.
func (r *CreateTopupInAppResponse) UnmarshalJSON(data []byte) error {
	type alias CreateTopupInAppResponse
	var raw struct {
		alias
		OrderID json.RawMessage `json:"order_id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = CreateTopupInAppResponse(raw.alias)
	r.OrderID = stringFromJSON(raw.OrderID)
	return nil
}

// ConfirmTopupInAppRequest confirms in-app top-up payment.
type ConfirmTopupInAppRequest struct {
	PaymentIntentID string `json:"payment_intent_id"`
	ClientSecret    string `json:"client_secret"`
}

// ListTopupPlans lists available top-up plans.
func (c *Client) ListTopupPlans(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[TopupPlan], *owlvigil.ResponseMeta, error) {
	var out ListResponse[TopupPlan]
	// The current Open API declares no pagination on this route.
	_ = opts
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/topup-plans", nil, nil, &out, reqOpts...)
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
	return c.ListTopupsWithFilters(
		ctx,
		OrderListOptions{Cursor: opts.Cursor, Limit: opts.Limit},
		reqOpts...,
	)
}

// ListTopupsWithFilters lists top-up transactions using every filter
// published by the current Open API.
func (c *Client) ListTopupsWithFilters(ctx context.Context, opts OrderListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Topup], *owlvigil.ResponseMeta, error) {
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
