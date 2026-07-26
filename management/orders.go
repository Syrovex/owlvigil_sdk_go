package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// OrderListOptions filters and paginates billing orders and top-ups.
type OrderListOptions struct {
	Cursor    string
	Limit     int
	OrderType string
	Status    string
}

func (o OrderListOptions) values() url.Values {
	q := ListOptions{Cursor: o.Cursor, Limit: o.Limit}.values()
	addFilter(q, "order_type", o.OrderType)
	addFilter(q, "status", o.Status)
	return q
}

// Order describes a billing order (subscription or top-up).
type Order struct {
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
	Balance               float64        `json:"balance,omitempty"`
	TotalRecharged        float64        `json:"total_recharged,omitempty"`

	// Legacy aliases retained for source compatibility.
	Type      string `json:"type,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

func (o *Order) UnmarshalJSON(data []byte) error {
	type alias Order
	var raw struct {
		alias
		ID      json.RawMessage `json:"id"`
		OrderID json.RawMessage `json:"order_id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*o = Order(raw.alias)
	o.ID = stringFromJSON(raw.ID)
	if o.ID == "" {
		o.ID = stringFromJSON(raw.OrderID)
	}
	if o.OrderType == "" {
		o.OrderType = o.Type
	}
	if o.Type == "" {
		o.Type = o.OrderType
	}
	if o.StripeSessionID == "" {
		o.StripeSessionID = o.SessionID
	}
	if o.SessionID == "" {
		o.SessionID = o.StripeSessionID
	}
	return nil
}

// ConfirmStripeSessionRequest confirms a Stripe checkout session.
type ConfirmStripeSessionRequest struct {
	SessionID string `json:"session_id,omitempty"`
}

// ListOrders lists billing orders.
func (c *Client) ListOrders(ctx context.Context, opts ListOptions, orderType string, reqOpts ...owlvigil.RequestOption) (*ListResponse[Order], *owlvigil.ResponseMeta, error) {
	return c.ListOrdersWithFilters(
		ctx,
		OrderListOptions{
			Cursor:    opts.Cursor,
			Limit:     opts.Limit,
			OrderType: orderType,
		},
		reqOpts...,
	)
}

// ListOrdersWithFilters lists billing orders using every filter published by
// the current Open API.
func (c *Client) ListOrdersWithFilters(ctx context.Context, opts OrderListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Order], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Order]
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/orders", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetOrder retrieves order details by ID.
func (c *Client) GetOrder(ctx context.Context, orderID string, reqOpts ...owlvigil.RequestOption) (*Order, *owlvigil.ResponseMeta, error) {
	var out Order
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/orders/"+url.PathEscape(orderID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ConfirmStripeSession confirms a Stripe checkout session and syncs order status.
func (c *Client) ConfirmStripeSession(ctx context.Context, orderID string, req *ConfirmStripeSessionRequest, reqOpts ...owlvigil.RequestOption) (*Order, *owlvigil.ResponseMeta, error) {
	var out Order
	path := "/billing/orders/" + url.PathEscape(orderID) + "/confirm-stripe-session"
	meta, err := c.http.Do(ctx, http.MethodPost, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
