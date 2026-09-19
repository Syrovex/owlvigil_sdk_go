package management

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// AutoChargePaymentMethod describes the card used for automatic charges.
type AutoChargePaymentMethod struct {
	Brand string `json:"brand"`
	Last4 string `json:"last4"`
}

// AutoCharge describes a workspace automatic-charge configuration and status.
type AutoCharge struct {
	Configured      bool                     `json:"configured"`
	WorkspaceID     int64                    `json:"workspace_id"`
	WorkspaceName   string                   `json:"workspace_name"`
	Enabled         bool                     `json:"enabled"`
	ThresholdAmount float64                  `json:"threshold_amount"`
	ChargeAmount    float64                  `json:"charge_amount"`
	Currency        string                   `json:"currency"`
	Status          string                   `json:"status"`
	PayerUserID     int64                    `json:"payer_user_id"`
	PayerName       string                   `json:"payer_name"`
	PaymentMethod   *AutoChargePaymentMethod `json:"payment_method,omitempty"`
	FailureReason   string                   `json:"failure_reason,omitempty"`
}

// UpdateAutoChargeRequest updates automatic charging for a workspace.
type UpdateAutoChargeRequest struct {
	WorkspaceID   int64   `json:"workspace_id"`
	Enabled       bool    `json:"enabled"`
	ChargeAmount  float64 `json:"charge_amount"`
	TransferPayer bool    `json:"transfer_payer,omitempty"`
}

// RetryAutoChargeRequest retries automatic charging for a workspace.
type RetryAutoChargeRequest struct {
	WorkspaceID int64 `json:"workspace_id"`
}

// GetAutoCharge retrieves automatic-charge configuration for a workspace.
func (c *Client) GetAutoCharge(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*AutoCharge, *owlvigil.ResponseMeta, error) {
	var out AutoCharge
	query := url.Values{"workspace_id": {strconv.FormatInt(workspaceID, 10)}}
	meta, err := c.http.Do(ctx, http.MethodGet, "/billing/auto-charge", query, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateAutoCharge updates automatic-charge configuration for a workspace.
func (c *Client) UpdateAutoCharge(ctx context.Context, req *UpdateAutoChargeRequest, reqOpts ...owlvigil.RequestOption) (*AutoCharge, *owlvigil.ResponseMeta, error) {
	var out AutoCharge
	meta, err := c.http.Do(ctx, http.MethodPut, "/billing/auto-charge", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// RetryAutoCharge retries a failed or currently eligible automatic charge.
func (c *Client) RetryAutoCharge(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*AutoCharge, *owlvigil.ResponseMeta, error) {
	var out AutoCharge
	req := RetryAutoChargeRequest{WorkspaceID: workspaceID}
	meta, err := c.http.Do(ctx, http.MethodPost, "/billing/auto-charge/retry", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
