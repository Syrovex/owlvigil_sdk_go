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
	ID                   string            `json:"id"`
	Slug                 string            `json:"slug,omitempty"`
	Name                 string            `json:"name"`
	Description          string            `json:"description,omitempty"`
	Price                float64           `json:"price"`
	PriceMonthly         float64           `json:"price_monthly,omitempty"`
	PriceYearly          *float64          `json:"price_yearly,omitempty"`
	Currency             string            `json:"currency"`
	Interval             string            `json:"interval"`
	StripePriceIDMonthly *string           `json:"stripe_price_id_monthly,omitempty"`
	StripePriceIDYearly  *string           `json:"stripe_price_id_yearly,omitempty"`
	Features             []string          `json:"features,omitempty"`
	FeatureText          string            `json:"feature_text,omitempty"`
	FeatureItems         []string          `json:"feature_items,omitempty"`
	Quotas               map[string]int64  `json:"quotas,omitempty"`
	MaxTeams             int               `json:"max_teams,omitempty"`
	MaxMembers           int               `json:"max_members,omitempty"`
	MaxAPIKeys           int               `json:"max_api_keys,omitempty"`
	MaxWorkspaces        int               `json:"max_workspaces,omitempty"`
	MonthlyRequestQuota  *float64          `json:"monthly_request_quota,omitempty"`
	MonthlyCostQuota     *float64          `json:"monthly_cost_quota,omitempty"`
	TrialDays            int               `json:"trial_days,omitempty"`
	Status               string            `json:"status,omitempty"`
	ForSale              bool              `json:"for_sale,omitempty"`
	IsPublic             bool              `json:"is_public,omitempty"`
	SortOrder            int               `json:"sort_order,omitempty"`
	CreatedAt            string            `json:"created_at,omitempty"`
	UpdatedAt            string            `json:"updated_at,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

// UnmarshalJSON accepts current and legacy plan feature representations.
func (p *Plan) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID                   json.RawMessage   `json:"id"`
		Slug                 string            `json:"slug,omitempty"`
		Name                 string            `json:"name"`
		Description          string            `json:"description,omitempty"`
		Price                float64           `json:"price"`
		PriceMonthly         float64           `json:"price_monthly,omitempty"`
		PriceYearly          *float64          `json:"price_yearly,omitempty"`
		Currency             string            `json:"currency"`
		Interval             string            `json:"interval"`
		StripePriceIDMonthly *string           `json:"stripe_price_id_monthly,omitempty"`
		StripePriceIDYearly  *string           `json:"stripe_price_id_yearly,omitempty"`
		Features             json.RawMessage   `json:"features,omitempty"`
		FeatureItems         []string          `json:"feature_items,omitempty"`
		Quotas               map[string]int64  `json:"quotas,omitempty"`
		MaxTeams             int               `json:"max_teams,omitempty"`
		MaxMembers           int               `json:"max_members,omitempty"`
		MaxAPIKeys           int               `json:"max_api_keys,omitempty"`
		MaxWorkspaces        int               `json:"max_workspaces,omitempty"`
		MonthlyRequestQuota  *float64          `json:"monthly_request_quota,omitempty"`
		MonthlyCostQuota     *float64          `json:"monthly_cost_quota,omitempty"`
		TrialDays            int               `json:"trial_days,omitempty"`
		Status               string            `json:"status,omitempty"`
		ForSale              bool              `json:"for_sale,omitempty"`
		IsPublic             bool              `json:"is_public,omitempty"`
		SortOrder            int               `json:"sort_order,omitempty"`
		CreatedAt            string            `json:"created_at,omitempty"`
		UpdatedAt            string            `json:"updated_at,omitempty"`
		Metadata             map[string]string `json:"metadata,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = Plan{
		ID:                   stringFromJSON(raw.ID),
		Slug:                 raw.Slug,
		Name:                 raw.Name,
		Description:          raw.Description,
		Price:                raw.Price,
		PriceMonthly:         raw.PriceMonthly,
		PriceYearly:          raw.PriceYearly,
		Currency:             raw.Currency,
		Interval:             raw.Interval,
		StripePriceIDMonthly: raw.StripePriceIDMonthly,
		StripePriceIDYearly:  raw.StripePriceIDYearly,
		FeatureItems:         raw.FeatureItems,
		Quotas:               raw.Quotas,
		MaxTeams:             raw.MaxTeams,
		MaxMembers:           raw.MaxMembers,
		MaxAPIKeys:           raw.MaxAPIKeys,
		MaxWorkspaces:        raw.MaxWorkspaces,
		MonthlyRequestQuota:  raw.MonthlyRequestQuota,
		MonthlyCostQuota:     raw.MonthlyCostQuota,
		TrialDays:            raw.TrialDays,
		Status:               raw.Status,
		ForSale:              raw.ForSale,
		IsPublic:             raw.IsPublic,
		SortOrder:            raw.SortOrder,
		CreatedAt:            raw.CreatedAt,
		UpdatedAt:            raw.UpdatedAt,
		Metadata:             raw.Metadata,
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
	UserID                        int    `json:"user_id"`
	SubscriptionPlanID            *int64 `json:"subscription_plan_id,omitempty"`
	SubscriptionStatus            string `json:"subscription_status"`
	SubscriptionTier              string `json:"subscription_tier"`
	EffectiveSubscriptionPlanID   *int64 `json:"effective_subscription_plan_id,omitempty"`
	EffectiveSubscriptionTier     string `json:"effective_subscription_tier"`
	StripeCustomerID              string `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID          string `json:"stripe_subscription_id,omitempty"`
	CurrentPeriodStart            string `json:"current_period_start,omitempty"`
	CurrentPeriodEnd              string `json:"current_period_end,omitempty"`
	MaxWorkspaces                 int    `json:"max_workspaces"`
	CurrentWorkspaceCount         int    `json:"current_workspace_count"`
	CanCreateMoreWorkspaces       bool   `json:"can_create_more_workspaces"`
	SubscriptionCancelAtPeriodEnd bool   `json:"subscription_cancel_at_period_end"`
	PendingPlanID                 *int64 `json:"pending_plan_id,omitempty"`
	PendingPlanEffectiveAt        string `json:"pending_plan_effective_at,omitempty"`
	StripeScheduleID              string `json:"stripe_schedule_id,omitempty"`
	TrialEndsAt                   string `json:"trial_ends_at,omitempty"`
	IsTrialExpired                bool   `json:"is_trial_expired"`
	Plan                          *Plan  `json:"plan,omitempty"`
	EffectivePlan                 *Plan  `json:"effective_plan,omitempty"`

	// Legacy aliases retained for source compatibility.
	ID                string  `json:"id,omitempty"`
	PlanID            string  `json:"plan_id,omitempty"`
	Status            string  `json:"status,omitempty"`
	CancelAtPeriodEnd bool    `json:"cancel_at_period_end,omitempty"`
	PendingDowngrade  *string `json:"pending_downgrade,omitempty"`
}

// UnmarshalJSON decodes the current subscription response and synchronizes
// the pre-refactor aliases.
func (s *Subscription) UnmarshalJSON(data []byte) error {
	type wireSubscription struct {
		UserID                        int             `json:"user_id"`
		SubscriptionPlanID            json.RawMessage `json:"subscription_plan_id"`
		SubscriptionStatus            string          `json:"subscription_status"`
		SubscriptionTier              string          `json:"subscription_tier"`
		EffectiveSubscriptionPlanID   json.RawMessage `json:"effective_subscription_plan_id"`
		EffectiveSubscriptionTier     string          `json:"effective_subscription_tier"`
		StripeCustomerID              string          `json:"stripe_customer_id"`
		StripeSubscriptionID          string          `json:"stripe_subscription_id"`
		CurrentPeriodStart            string          `json:"current_period_start"`
		CurrentPeriodEnd              string          `json:"current_period_end"`
		MaxWorkspaces                 int             `json:"max_workspaces"`
		CurrentWorkspaceCount         int             `json:"current_workspace_count"`
		CanCreateMoreWorkspaces       bool            `json:"can_create_more_workspaces"`
		SubscriptionCancelAtPeriodEnd bool            `json:"subscription_cancel_at_period_end"`
		PendingPlanID                 json.RawMessage `json:"pending_plan_id"`
		PendingPlanEffectiveAt        string          `json:"pending_plan_effective_at"`
		StripeScheduleID              string          `json:"stripe_schedule_id"`
		TrialEndsAt                   string          `json:"trial_ends_at"`
		IsTrialExpired                bool            `json:"is_trial_expired"`
		Plan                          *Plan           `json:"plan"`
		EffectivePlan                 *Plan           `json:"effective_plan"`
		ID                            json.RawMessage `json:"id"`
		PlanID                        json.RawMessage `json:"plan_id"`
		Status                        string          `json:"status"`
		CancelAtPeriodEnd             bool            `json:"cancel_at_period_end"`
		PendingDowngrade              *string         `json:"pending_downgrade"`
	}
	var raw wireSubscription
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*s = Subscription{
		UserID:                        raw.UserID,
		SubscriptionPlanID:            int64PointerFromJSON(raw.SubscriptionPlanID),
		SubscriptionStatus:            raw.SubscriptionStatus,
		SubscriptionTier:              raw.SubscriptionTier,
		EffectiveSubscriptionPlanID:   int64PointerFromJSON(raw.EffectiveSubscriptionPlanID),
		EffectiveSubscriptionTier:     raw.EffectiveSubscriptionTier,
		StripeCustomerID:              raw.StripeCustomerID,
		StripeSubscriptionID:          raw.StripeSubscriptionID,
		CurrentPeriodStart:            raw.CurrentPeriodStart,
		CurrentPeriodEnd:              raw.CurrentPeriodEnd,
		MaxWorkspaces:                 raw.MaxWorkspaces,
		CurrentWorkspaceCount:         raw.CurrentWorkspaceCount,
		CanCreateMoreWorkspaces:       raw.CanCreateMoreWorkspaces,
		SubscriptionCancelAtPeriodEnd: raw.SubscriptionCancelAtPeriodEnd,
		PendingPlanID:                 int64PointerFromJSON(raw.PendingPlanID),
		PendingPlanEffectiveAt:        raw.PendingPlanEffectiveAt,
		StripeScheduleID:              raw.StripeScheduleID,
		TrialEndsAt:                   raw.TrialEndsAt,
		IsTrialExpired:                raw.IsTrialExpired,
		Plan:                          raw.Plan,
		EffectivePlan:                 raw.EffectivePlan,
		ID:                            stringFromJSON(raw.ID),
		PlanID:                        stringFromJSON(raw.PlanID),
		Status:                        raw.Status,
		CancelAtPeriodEnd:             raw.CancelAtPeriodEnd,
		PendingDowngrade:              raw.PendingDowngrade,
	}
	if s.PlanID == "" {
		s.PlanID = stringFromJSON(raw.SubscriptionPlanID)
	}
	if s.Status == "" {
		s.Status = s.SubscriptionStatus
	}
	if s.SubscriptionStatus == "" {
		s.SubscriptionStatus = s.Status
	}
	if !s.CancelAtPeriodEnd {
		s.CancelAtPeriodEnd = s.SubscriptionCancelAtPeriodEnd
	}
	if !s.SubscriptionCancelAtPeriodEnd {
		s.SubscriptionCancelAtPeriodEnd = s.CancelAtPeriodEnd
	}
	if s.PendingDowngrade == nil && len(raw.PendingPlanID) > 0 && string(raw.PendingPlanID) != "null" {
		value := stringFromJSON(raw.PendingPlanID)
		s.PendingDowngrade = &value
	}
	return nil
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
	CheckoutURL string         `json:"checkout_url"`
	SessionID   string         `json:"session_id"`
	OrderID     string         `json:"order_id,omitempty"`
	Status      string         `json:"status,omitempty"`
	ExpiresAt   string         `json:"expires_at,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UnmarshalJSON accepts numeric and string order IDs.
func (r *CreateSubscriptionCheckoutResponse) UnmarshalJSON(data []byte) error {
	type alias CreateSubscriptionCheckoutResponse
	var raw struct {
		alias
		OrderID json.RawMessage `json:"order_id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = CreateSubscriptionCheckoutResponse(raw.alias)
	r.OrderID = stringFromJSON(raw.OrderID)
	return nil
}

// CreateSubscriptionInAppRequest creates an in-app subscription.
type CreateSubscriptionInAppRequest struct {
	PlanID    string `json:"plan_id"`
	Interval  string `json:"interval"`
	ReturnURL string `json:"return_url"`
}

// CreateSubscriptionInAppResponse contains payment intent details.
type CreateSubscriptionInAppResponse struct {
	ClientSecret         string  `json:"client_secret"`
	StripeSubscriptionID string  `json:"stripe_subscription_id,omitempty"`
	PaymentIntentID      string  `json:"payment_intent_id"`
	Status               string  `json:"status"`
	Amount               float64 `json:"amount,omitempty"`
	Currency             string  `json:"currency,omitempty"`

	// RequiresAction is retained for pre-refactor responses.
	RequiresAction bool `json:"requires_action,omitempty"`
}

// ConfirmSubscriptionInAppRequest confirms in-app subscription payment.
type ConfirmSubscriptionInAppRequest struct {
	PlanID               string `json:"plan_id"`
	StripeSubscriptionID string `json:"stripe_subscription_id"`

	// PaymentIntentID is retained for source compatibility with SDK versions
	// that predate the current Open API contract. It is not sent on the wire.
	PaymentIntentID string `json:"-"`
}

// UpgradeSubscriptionRequest upgrades subscription plan.
type UpgradeSubscriptionRequest struct {
	PlanID   string `json:"plan_id"`
	Interval string `json:"interval"`

	// NewPlanID is the legacy alias for PlanID.
	NewPlanID string `json:"-"`
}

// DowngradeSubscriptionRequest downgrades subscription plan.
type DowngradeSubscriptionRequest struct {
	PlanID   string `json:"plan_id"`
	Interval string `json:"interval"`

	// NewPlanID is the legacy alias for PlanID.
	NewPlanID string `json:"-"`
}

// CancelSubscriptionRequest controls how a subscription is canceled.
type CancelSubscriptionRequest struct {
	Reason            string `json:"reason,omitempty"`
	CancelImmediately bool   `json:"cancel_immediately,omitempty"`
}

// MarshalJSON emits the current Open API confirmation contract.
func (r ConfirmSubscriptionInAppRequest) MarshalJSON() ([]byte, error) {
	type wireRequest struct {
		PlanID               string `json:"plan_id"`
		StripeSubscriptionID string `json:"stripe_subscription_id"`
	}
	return json.Marshal(wireRequest{
		PlanID:               r.PlanID,
		StripeSubscriptionID: r.StripeSubscriptionID,
	})
}

// MarshalJSON emits the current Open API change-subscription contract.
func (r UpgradeSubscriptionRequest) MarshalJSON() ([]byte, error) {
	planID := r.PlanID
	if planID == "" {
		planID = r.NewPlanID
	}
	return json.Marshal(struct {
		PlanID   string `json:"plan_id"`
		Interval string `json:"interval"`
	}{
		PlanID:   planID,
		Interval: r.Interval,
	})
}

// MarshalJSON emits the current Open API change-subscription contract.
func (r DowngradeSubscriptionRequest) MarshalJSON() ([]byte, error) {
	planID := r.PlanID
	if planID == "" {
		planID = r.NewPlanID
	}
	return json.Marshal(struct {
		PlanID   string `json:"plan_id"`
		Interval string `json:"interval"`
	}{
		PlanID:   planID,
		Interval: r.Interval,
	})
}

// CheckoutSession describes a checkout session.
type CheckoutSession struct {
	SessionID   string         `json:"session_id"`
	CheckoutURL string         `json:"checkout_url,omitempty"`
	OrderID     string         `json:"order_id,omitempty"`
	Status      string         `json:"status"`
	ExpiresAt   string         `json:"expires_at,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UnmarshalJSON accepts numeric and string order IDs.
func (s *CheckoutSession) UnmarshalJSON(data []byte) error {
	type alias CheckoutSession
	var raw struct {
		alias
		OrderID json.RawMessage `json:"order_id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*s = CheckoutSession(raw.alias)
	s.OrderID = stringFromJSON(raw.OrderID)
	return nil
}

// ListPlans lists available subscription plans.
func (c *Client) ListPlans(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Plan], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Plan]
	// The current Open API declares no pagination on this route.
	_ = opts
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/plans", nil, nil, &out, reqOpts...)
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
	return c.CancelSubscriptionWithRequest(ctx, &CancelSubscriptionRequest{}, reqOpts...)
}

// CancelSubscriptionWithRequest cancels a subscription with the optional
// reason and immediate-cancellation behavior supported by the Open API.
func (c *Client) CancelSubscriptionWithRequest(ctx context.Context, req *CancelSubscriptionRequest, reqOpts ...owlvigil.RequestOption) (*Subscription, *owlvigil.ResponseMeta, error) {
	var out Subscription
	body := any(req)
	if req == nil {
		body = struct{}{}
	}
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/subscription/cancel", nil, body, &out, reqOpts...)
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
