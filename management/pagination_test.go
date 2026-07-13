package management

import (
	"net/url"
	"testing"
)

func TestListOptionsValues(t *testing.T) {
	t.Parallel()

	values := ListOptions{Cursor: "cur_1", Limit: 25}.values()

	if values.Get("cursor") != "cur_1" || values.Get("limit") != "25" {
		t.Fatalf("values = %s", values.Encode())
	}
}

func TestAddFilterEmptyValue(t *testing.T) {
	t.Parallel()

	values := url.Values{"existing": []string{"value"}}
	got := addFilter(values, "gateway_key_id", "")
	if got.Get("existing") != "value" {
		t.Fatalf("values = %s", got.Encode())
	}
	if got.Get("gateway_key_id") != "" {
		t.Fatalf("gateway_key_id = %q", got.Get("gateway_key_id"))
	}

	got = addFilter(values, "gateway_key_id", "9")
	if got.Get("gateway_key_id") != "9" {
		t.Fatalf("gateway_key_id = %q", got.Get("gateway_key_id"))
	}
}
