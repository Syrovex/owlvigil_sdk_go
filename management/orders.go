package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// Order describes a billing order (subscription or top-up).
type Order struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at,omitempty"`
	SessionID string  `json:"session_id,omitempty"`
}

func (o *Order) UnmarshalJSON(data []byte) error {
	type alias Order
	var raw struct {
		alias
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*o = Order(raw.alias)
	o.ID = stringFromJSON(raw.ID)
	return nil
}

// ConfirmStripeSessionRequest confirms a Stripe checkout session.
type ConfirmStripeSessionRequest struct {
	SessionID string `json:"session_id"`
}

// ListOrders lists billing orders.
func (c *Client) ListOrders(ctx context.Context, opts ListOptions, orderType string, reqOpts ...owlvigil.RequestOption) (*ListResponse[Order], *owlvigil.ResponseMeta, error) {
	q := addFilter(opts.values(), "order_type", orderType)
	var out ListResponse[Order]
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/orders", q, nil, &out, reqOpts...)
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
