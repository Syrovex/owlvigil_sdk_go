package management

import (
	"encoding/json"
	"net/url"
)

// ListOptions contains common cursor pagination options.
type ListOptions struct {
	Cursor string
	Limit  int
}

// PageInfo contains cursor pagination metadata.
type PageInfo struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more,omitempty"`
}

// ListResponse contains typed list items and pagination metadata.
type ListResponse[T any] struct {
	Items       []T               `json:"items"`
	Groups      []PermissionGroup `json:"groups,omitempty"`
	InviteRoles []T               `json:"invite_roles,omitempty"`
	EditRoles   []T               `json:"edit_roles,omitempty"`
	PageInfo    PageInfo          `json:"page_info"`
}

func (r *ListResponse[T]) UnmarshalJSON(data []byte) error {
	var items []T
	if err := json.Unmarshal(data, &items); err == nil {
		r.Items = items
		return nil
	}
	type alias ListResponse[T]
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*r = ListResponse[T](out)
	return nil
}

func (o ListOptions) values() url.Values {
	q := url.Values{}
	if o.Cursor != "" {
		q.Set("cursor", o.Cursor)
	}
	if o.Limit > 0 {
		q.Set("limit", intString(o.Limit))
	}
	return q
}
