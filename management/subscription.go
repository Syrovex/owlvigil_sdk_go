package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// Plan describes a subscription plan.
type Plan struct {
	ID            string            `json:"id"`
	Slug          string            `json:"slug,omitempty"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Price         float64           `json:"price"`
	PriceMonthly  float64           `json:"price_monthly,omitempty"`
	PriceYearly   *float64          `json:"price_yearly,omitempty"`
	Currency      string            `json:"currency"`
	Interval      string            `json:"interval"`
	Features      []string          `json:"features,omitempty"`
	FeatureText   string            `json:"feature_text,omitempty"`
	FeatureItems  []string          `json:"feature_items,omitempty"`
	Quotas        map[string]int64  `json:"quotas,omitempty"`
	MaxTeams      int               `json:"max_teams,omitempty"`
	MaxMembers    int               `json:"max_members,omitempty"`
	MaxAPIKeys    int               `json:"max_api_keys,omitempty"`
	MaxWorkspaces int               `json:"max_workspaces,omitempty"`
	TrialDays     int               `json:"trial_days,omitempty"`
	Status        string            `json:"status,omitempty"`
	ForSale       bool              `json:"for_sale,omitempty"`
	IsPublic      bool              `json:"is_public,omitempty"`
	SortOrder     int               `json:"sort_order,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

func (p *Plan) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID            json.RawMessage   `json:"id"`
		Slug          string            `json:"slug,omitempty"`
		Name          string            `json:"name"`
		Description   string            `json:"description,omitempty"`
		Price         float64           `json:"price"`
		PriceMonthly  float64           `json:"price_monthly,omitempty"`
		PriceYearly   *float64          `json:"price_yearly,omitempty"`
		Currency      string            `json:"currency"`
		Interval      string            `json:"interval"`
		Features      json.RawMessage   `json:"features,omitempty"`
		FeatureItems  []string          `json:"feature_items,omitempty"`
		Quotas        map[string]int64  `json:"quotas,omitempty"`
		MaxTeams      int               `json:"max_teams,omitempty"`
		MaxMembers    int               `json:"max_members,omitempty"`
		MaxAPIKeys    int               `json:"max_api_keys,omitempty"`
		MaxWorkspaces int               `json:"max_workspaces,omitempty"`
		TrialDays     int               `json:"trial_days,omitempty"`
		Status        string            `json:"status,omitempty"`
		ForSale       bool              `json:"for_sale,omitempty"`
		IsPublic      bool              `json:"is_public,omitempty"`
		SortOrder     int               `json:"sort_order,omitempty"`
		Metadata      map[string]string `json:"metadata,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = Plan{
		ID:            stringFromJSON(raw.ID),
		Slug:          raw.Slug,
		Name:          raw.Name,
		Description:   raw.Description,
		Price:         raw.Price,
		PriceMonthly:  raw.PriceMonthly,
		PriceYearly:   raw.PriceYearly,
		Currency:      raw.Currency,
		Interval:      raw.Interval,
		FeatureItems:  raw.FeatureItems,
		Quotas:        raw.Quotas,
		MaxTeams:      raw.MaxTeams,
		MaxMembers:    raw.MaxMembers,
		MaxAPIKeys:    raw.MaxAPIKeys,
		MaxWorkspaces: raw.MaxWorkspaces,
		TrialDays:     raw.TrialDays,
		Status:        raw.Status,
		ForSale:       raw.ForSale,
		IsPublic:      raw.IsPublic,
		SortOrder:     raw.SortOrder,
		Metadata:      raw.Metadata,
	}
	p.ID = stringFromJSON(raw.ID)
	if p.ID == "" {
		p.ID = p.Slug
	}
	if len(raw.Features) > 0 && string(raw.Features) != "null" {
		var items []string
		if err := json.Unmarshal(raw.Features, &items); err == nil {
			p.Features = items
		} else {
			var text string
			if err := json.Unmarshal(raw.Features, &text); err == nil {
				p.FeatureText = text
				p.Features = splitFeatureText(text)
			}
		}
	}
	if len(p.Features) == 0 && len(p.FeatureItems) > 0 {
		p.Features = p.FeatureItems
	}
	return nil
}

func splitFeatureText(text string) []string {
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == '\n' || r == ',' || r == ';'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// Subscription describes a user or workspace subscription.
type Subscription struct {
	ID                string  `json:"id"`
	PlanID            string  `json:"plan_id"`
	Status            string  `json:"status"`
	CurrentPeriodEnd  string  `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd bool    `json:"cancel_at_period_end"`
	PendingDowngrade  *string `json:"pending_downgrade,omitempty"`
}

// CreateSubscriptionCheckoutRequest creates a subscription checkout session.
type CreateSubscriptionCheckoutRequest struct {
	PlanID     string `json:"plan_id"`
	Interval   string `json:"interval"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// CreateSubscriptionCheckoutResponse contains checkout session details.
type CreateSubscriptionCheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
	OrderID     string `json:"order_id,omitempty"`
}

// CreateSubscriptionInAppRequest creates an in-app subscription.
type CreateSubscriptionInAppRequest struct {
	PlanID   string `json:"plan_id"`
	Interval string `json:"interval"`
}

// CreateSubscriptionInAppResponse contains payment intent details.
type CreateSubscriptionInAppResponse struct {
	PaymentIntentID string `json:"payment_intent_id"`
	ClientSecret    string `json:"client_secret"`
	RequiresAction  bool   `json:"requires_action"`
	Status          string `json:"status"`
}

// ConfirmSubscriptionInAppRequest confirms in-app subscription payment.
type ConfirmSubscriptionInAppRequest struct {
	PaymentIntentID string `json:"payment_intent_id"`
}

// UpgradeSubscriptionRequest upgrades subscription plan.
type UpgradeSubscriptionRequest struct {
	NewPlanID string `json:"new_plan_id"`
}

// DowngradeSubscriptionRequest downgrades subscription plan.
type DowngradeSubscriptionRequest struct {
	NewPlanID string `json:"new_plan_id"`
}

// CheckoutSession describes a checkout session.
type CheckoutSession struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	OrderID   string `json:"order_id,omitempty"`
}

// ListPlans lists available subscription plans.
func (c *Client) ListPlans(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Plan], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Plan]
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/plans", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetPlan retrieves plan details by ID.
func (c *Client) GetPlan(ctx context.Context, planID string, reqOpts ...owlvigil.RequestOption) (*Plan, *owlvigil.ResponseMeta, error) {
	var out Plan
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/plans/"+url.PathEscape(planID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetSubscription retrieves current subscription status.
func (c *Client) GetSubscription(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*Subscription, *owlvigil.ResponseMeta, error) {
	var out Subscription
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/subscription", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateSubscriptionCheckout creates a subscription checkout session.
func (c *Client) CreateSubscriptionCheckout(ctx context.Context, req *CreateSubscriptionCheckoutRequest, reqOpts ...owlvigil.RequestOption) (*CreateSubscriptionCheckoutResponse, *owlvigil.ResponseMeta, error) {
	var out CreateSubscriptionCheckoutResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/subscription/checkout", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateSubscriptionInApp creates an in-app subscription payment.
func (c *Client) CreateSubscriptionInApp(ctx context.Context, req *CreateSubscriptionInAppRequest, reqOpts ...owlvigil.RequestOption) (*CreateSubscriptionInAppResponse, *owlvigil.ResponseMeta, error) {
	var out CreateSubscriptionInAppResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/subscription/in-app", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ConfirmSubscriptionInApp confirms in-app subscription payment after 3DS.
func (c *Client) ConfirmSubscriptionInApp(ctx context.Context, req *ConfirmSubscriptionInAppRequest, reqOpts ...owlvigil.RequestOption) (*Subscription, *owlvigil.ResponseMeta, error) {
	var out Subscription
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/subscription/in-app/confirm", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpgradeSubscription upgrades subscription to a higher plan.
func (c *Client) UpgradeSubscription(ctx context.Context, req *UpgradeSubscriptionRequest, reqOpts ...owlvigil.RequestOption) (*Subscription, *owlvigil.ResponseMeta, error) {
	var out Subscription
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/subscription/upgrade", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DowngradeSubscription schedules subscription downgrade at period end.
func (c *Client) DowngradeSubscription(ctx context.Context, req *DowngradeSubscriptionRequest, reqOpts ...owlvigil.RequestOption) (*Subscription, *owlvigil.ResponseMeta, error) {
	var out Subscription
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/subscription/downgrade", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CancelSubscription cancels subscription at period end.
func (c *Client) CancelSubscription(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*Subscription, *owlvigil.ResponseMeta, error) {
	var out Subscription
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/subscription/cancel", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ReactivateSubscription reactivates a canceled subscription.
func (c *Client) ReactivateSubscription(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*Subscription, *owlvigil.ResponseMeta, error) {
	var out Subscription
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/subscription/reactivate", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetSubscriptionCheckoutSession retrieves checkout session status.
func (c *Client) GetSubscriptionCheckoutSession(ctx context.Context, sessionID string, reqOpts ...owlvigil.RequestOption) (*CheckoutSession, *owlvigil.ResponseMeta, error) {
	var out CheckoutSession
	path := "/billing/subscription/checkout-sessions/" + url.PathEscape(sessionID)
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// SyncLatestSubscriptionCheckout syncs the latest subscription checkout session.
func (c *Client) SyncLatestSubscriptionCheckout(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*CheckoutSession, *owlvigil.ResponseMeta, error) {
	var out CheckoutSession
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/subscription/checkout-sessions/sync-latest", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
