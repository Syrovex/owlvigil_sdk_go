package management

import (
	"context"
	"net/http"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// Team describes a workspace team.
type Team struct {
	ID                 int64    `json:"id"`
	WorkspaceID        int64    `json:"workspace_id"`
	ProjectID          int      `json:"project_id,omitempty"`
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	Status             string   `json:"status"`
	MemberCount        int      `json:"member_count,omitempty"`
	MonthlyBudgetLimit *float64 `json:"monthly_budget_limit,omitempty"`
	IsDefault          bool     `json:"is_default,omitempty"`
	CreatedAt          string   `json:"created_at,omitempty"`
	UpdatedAt          string   `json:"updated_at,omitempty"`
	DeletedAt          *string  `json:"deleted_at,omitempty"`
}

// CreateTeamRequest creates a team.
type CreateTeamRequest struct {
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	MonthlyBudgetLimit *float64 `json:"monthly_budget_limit,omitempty"`
}

// UpdateTeamRequest updates team fields.
type UpdateTeamRequest struct {
	Name               *string  `json:"name,omitempty"`
	Description        *string  `json:"description,omitempty"`
	MonthlyBudgetLimit *float64 `json:"monthly_budget_limit,omitempty"`

	// Status is retained for source compatibility and is not mutable in the current Open API.
	Status *string `json:"-"`
}

// ListTeams lists teams in a workspace.
func (c *Client) ListTeams(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Team], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Team]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/teams"
	meta, err := c.http.Do(ctx, http.MethodGet, path, opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateTeam creates a new team.
func (c *Client) CreateTeam(ctx context.Context, workspaceID int64, req *CreateTeamRequest, reqOpts ...owlvigil.RequestOption) (*Team, *owlvigil.ResponseMeta, error) {
	var out Team
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/teams"
	meta, err := c.http.Do(ctx, http.MethodPost, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetTeam retrieves team details by ID.
func (c *Client) GetTeam(ctx context.Context, workspaceID, teamID int64, reqOpts ...owlvigil.RequestOption) (*Team, *owlvigil.ResponseMeta, error) {
	var out Team
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/teams/" + strconv.FormatInt(teamID, 10)
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateTeam updates team information.
func (c *Client) UpdateTeam(ctx context.Context, workspaceID, teamID int64, req *UpdateTeamRequest, reqOpts ...owlvigil.RequestOption) (*Team, *owlvigil.ResponseMeta, error) {
	var out Team
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/teams/" + strconv.FormatInt(teamID, 10)
	meta, err := c.http.Do(ctx, http.MethodPatch, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeleteTeam archives a team.
func (c *Client) DeleteTeam(ctx context.Context, workspaceID, teamID int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.DeleteTeamWithResult(ctx, workspaceID, teamID, reqOpts...)
	return meta, err
}

// DeleteTeamWithResult archives a team and returns confirmation.
func (c *Client) DeleteTeamWithResult(ctx context.Context, workspaceID, teamID int64, reqOpts ...owlvigil.RequestOption) (*DeleteResponse, *owlvigil.ResponseMeta, error) {
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/teams/" + strconv.FormatInt(teamID, 10)
	var out DeleteResponse
	meta, err := c.http.Do(ctx, http.MethodDelete, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
