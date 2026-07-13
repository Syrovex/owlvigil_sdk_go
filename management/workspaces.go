package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// Workspace describes an OwlVigil workspace visible to the user.
type Workspace struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
	Plan        string `json:"plan,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// UpdateWorkspaceRequest updates workspace fields.
type UpdateWorkspaceRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ActivityRecord describes a workspace activity log entry.
type ActivityRecord struct {
	ID          string `json:"id"`
	WorkspaceID int64  `json:"workspace_id"`
	ActorID     int64  `json:"actor_id"`
	ActorEmail  string `json:"actor_email,omitempty"`
	Action      string `json:"action"`
	Resource    string `json:"resource"`
	ResourceID  string `json:"resource_id,omitempty"`
	Details     any    `json:"details,omitempty"`
	Timestamp   string `json:"timestamp"`
	Tab         string `json:"tab,omitempty"`
	Who         string `json:"who,omitempty"`
	What        string `json:"what,omitempty"`
	Meta        string `json:"meta,omitempty"`
	IP          string `json:"ip,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

func (a *ActivityRecord) UnmarshalJSON(data []byte) error {
	type alias ActivityRecord
	var raw struct {
		alias
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*a = ActivityRecord(raw.alias)
	a.ID = stringFromJSON(raw.ID)
	if a.Action == "" {
		a.Action = a.What
	}
	if a.Timestamp == "" {
		a.Timestamp = a.CreatedAt
	}
	return nil
}

// ListWorkspaces lists workspaces visible to the authenticated user.
func (c *Client) ListWorkspaces(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Workspace], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Workspace]
	meta, err := c.http.Do(ctx, http.MethodGet, "/workspaces", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetWorkspace retrieves a workspace by ID.
func (c *Client) GetWorkspace(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*Workspace, *owlvigil.ResponseMeta, error) {
	var out Workspace
	meta, err := c.http.Do(ctx, http.MethodGet, "/workspaces/"+strconv.FormatInt(id, 10), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateWorkspace updates workspace information.
func (c *Client) UpdateWorkspace(ctx context.Context, id int64, req *UpdateWorkspaceRequest, reqOpts ...owlvigil.RequestOption) (*Workspace, *owlvigil.ResponseMeta, error) {
	var out Workspace
	meta, err := c.http.Do(ctx, http.MethodPatch, "/workspaces/"+strconv.FormatInt(id, 10), nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListWorkspaceActivity lists workspace activity logs.
func (c *Client) ListWorkspaceActivity(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[ActivityRecord], *owlvigil.ResponseMeta, error) {
	var out ListResponse[ActivityRecord]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/activity"
	meta, err := c.http.Do(ctx, http.MethodGet, path, opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

func intString(i int) string {
	return strconv.Itoa(i)
}

func addFilter(q url.Values, key, value string) url.Values {
	if value != "" {
		q.Set(key, value)
	}
	return q
}
