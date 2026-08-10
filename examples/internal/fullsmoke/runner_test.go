package fullsmoke

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

func TestSmokeUseCasesMatchManagementOperationCatalog(t *testing.T) {
	const (
		wantOperationCount = 141
		// SHA-256 of the sorted Dashboard/Open API Management route catalog.
		wantCatalogHash = "e4ee307fef2b3687f552b67d0881303b6454e7f45de2fcf3cd045d8472b91b72"
	)

	source, err := os.ReadFile("runner.go")
	if err != nil {
		t.Fatalf("os.ReadFile(runner.go) error = %v", err)
	}
	pattern := regexp.MustCompile(`"(DELETE|GET|PATCH|POST|PUT) (/v1/[^"]+)"`)
	matches := pattern.FindAllStringSubmatch(string(source), -1)
	unique := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		unique[match[1]+" "+match[2]] = struct{}{}
	}
	operations := make([]string, 0, len(unique))
	for operation := range unique {
		operations = append(operations, operation)
	}
	sort.Strings(operations)

	if len(operations) != wantOperationCount {
		t.Errorf("openapi-smoke Management operation count = %d, want %d", len(operations), wantOperationCount)
	}
	gotHash := sha256.Sum256([]byte(strings.Join(operations, "\n") + "\n"))
	if got := hexString(gotHash[:]); got != wantCatalogHash {
		t.Errorf("openapi-smoke Management operation catalog SHA-256 = %s, want %s", got, wantCatalogHash)
	}
}

func hexString(value []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, len(value)*2)
	for index, item := range value {
		out[index*2] = digits[item>>4]
		out[index*2+1] = digits[item&0x0f]
	}
	return string(out)
}

func TestOAuthEnabled(t *testing.T) {
	tests := []struct {
		name         string
		accessToken  string
		clientID     string
		clientSecret string
		want         bool
	}{
		{name: "access token", accessToken: "token", want: true},
		{name: "client credentials", clientID: "client", clientSecret: "secret", want: true},
		{name: "no OAuth credentials", want: false},
		{name: "client ID only", clientID: "client", want: false},
		{name: "client secret only", clientSecret: "secret", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := oauthEnabled(tt.accessToken, tt.clientID, tt.clientSecret); got != tt.want {
				t.Errorf("oauthEnabled(%q, %q, %q) = %t, want %t", tt.accessToken, tt.clientID, tt.clientSecret, got, tt.want)
			}
		})
	}
}

func TestWriteSmokeEnabled(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "1", want: true},
		{value: "true", want: false},
		{value: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := writeSmokeEnabled(tt.value); got != tt.want {
				t.Errorf("writeSmokeEnabled(%q) = %t, want %t", tt.value, got, tt.want)
			}
		})
	}
}

func TestStripePaymentMethodID(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "pm_card_visa", want: "pm_card_visa"},
		{value: " pm_card_visa ", want: "pm_card_visa"},
		{value: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := stripePaymentMethodID(tt.value); got != tt.want {
				t.Errorf("stripePaymentMethodID(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestFirstSaleableMonthlyPlan(t *testing.T) {
	monthlyPriceID := "price_team_monthly"
	emptyPriceID := "  "
	tests := []struct {
		name         string
		plans        []management.Plan
		wantPlanID   string
		wantInterval string
	}{
		{
			name: "selects first self-service monthly plan",
			plans: []management.Plan{
				{ID: "1", Name: "Individual", ForSale: true},
				{ID: "2", Name: "Team", ForSale: true, StripePriceIDMonthly: &monthlyPriceID},
				{ID: "7", Name: "Business", ForSale: true, StripePriceIDMonthly: &monthlyPriceID},
			},
			wantPlanID:   "2",
			wantInterval: "monthly",
		},
		{
			name: "ignores plans without a usable monthly price",
			plans: []management.Plan{
				{ID: "1", Name: "Individual", ForSale: true},
				{ID: "5", Name: "Enterprise", ForSale: true, StripePriceIDMonthly: &emptyPriceID},
				{ID: "7", Name: "Business", ForSale: false, StripePriceIDMonthly: &monthlyPriceID},
				{ID: "", Name: "Missing ID", ForSale: true, StripePriceIDMonthly: &monthlyPriceID},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPlanID, gotInterval := firstSaleableMonthlyPlan(tt.plans)
			if gotPlanID != tt.wantPlanID || gotInterval != tt.wantInterval {
				t.Errorf(
					"firstSaleableMonthlyPlan(%+v) = (%q, %q), want (%q, %q)",
					tt.plans,
					gotPlanID,
					gotInterval,
					tt.wantPlanID,
					tt.wantInterval,
				)
			}
		})
	}
}

func TestPositiveAmount(t *testing.T) {
	tests := []struct {
		value   string
		want    float64
		wantErr bool
	}{
		{value: "10.50", want: 10.5},
		{value: " 1 ", want: 1},
		{value: "", wantErr: true},
		{value: "0", wantErr: true},
		{value: "-1", wantErr: true},
		{value: "NaN", wantErr: true},
		{value: "+Inf", wantErr: true},
		{value: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := positiveAmount(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("positiveAmount(%q) error = %v, wantErr %t", tt.value, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("positiveAmount(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestPositiveLimitOrDefault(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{name: "positive", value: 42, want: 42},
		{name: "zero", value: 0, want: 1},
		{name: "negative", value: -1, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := positiveLimitOrDefault(tt.value); got != tt.want {
				t.Errorf("positiveLimitOrDefault(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestFirstRegisteredMemberUserID(t *testing.T) {
	members := []management.Member{
		{ID: 10, UserID: 0, Email: "pending@example.com"},
		{ID: 11, UserID: 64, Email: "owner@example.com"},
	}
	if got := firstRegisteredMemberUserID(members); got != 64 {
		t.Errorf("firstRegisteredMemberUserID() = %d, want 64", got)
	}
}

func TestCurrentUserProfileUpdateRequestUsesUsername(t *testing.T) {
	profile := &management.UserProfile{
		Username:           "sdk-smoke-user",
		Name:               "shared display name",
		AvatarURL:          "https://example.com/avatar.png",
		DefaultWorkspaceID: 295,
	}

	body, err := json.Marshal(currentUserProfileUpdateRequest(profile))
	if err != nil {
		t.Fatalf("marshal profile update request: %v", err)
	}
	var requestBody map[string]any
	if err := json.Unmarshal(body, &requestBody); err != nil {
		t.Fatalf("unmarshal profile update request: %v", err)
	}
	if got := requestBody["username"]; got != profile.Username {
		t.Errorf("currentUserProfileUpdateRequest(%+v) username = %v, want %q", profile, got, profile.Username)
	}
	if got := requestBody["username"]; got == profile.Name {
		t.Errorf("currentUserProfileUpdateRequest(%+v) username = %v, want account username %q", profile, got, profile.Username)
	}
}

func TestFirstRetryableWebhookEventID(t *testing.T) {
	const endpointID int64 = 101
	events := []management.WebhookEvent{
		{ID: "other-endpoint", EndpointID: 202, Status: "failed", Attempts: 1},
		{ID: "delivered", EndpointID: endpointID, Status: "delivered", Attempts: 1},
		{ID: "exhausted", EndpointID: endpointID, Status: "failed", Attempts: 5},
		{ID: "retryable", EndpointID: endpointID, Status: "DEAD", Attempts: 4},
		{ID: "later", EndpointID: endpointID, Status: "failed", Attempts: 1},
	}

	if got, ok := firstRetryableWebhookEventID(events, endpointID); !ok || got != "retryable" {
		t.Errorf("firstRetryableWebhookEventID(%+v, %d) = (%q, %t), want (%q, true)", events, endpointID, got, ok, "retryable")
	}
	if got, ok := firstRetryableWebhookEventID(events[:3], endpointID); ok || got != "" {
		t.Errorf("firstRetryableWebhookEventID(%+v, %d) = (%q, %t), want (\"\", false)", events[:3], endpointID, got, ok)
	}
}

func TestRunner_CallSkipKnownRequiresMatchingAPIError(t *testing.T) {
	expected := expectedUpgradeRequired("feature.audit_logs is not included")
	tests := []struct {
		name       string
		err        error
		wantStatus string
	}{
		{
			name: "matching API error is skipped",
			err: &owlvigil.APIError{
				StatusCode: http.StatusPaymentRequired,
				Code:       "upgrade_required",
				Message:    "feature.audit_logs is not included in the Team plan",
			},
			wantStatus: "SKIP",
		},
		{
			name: "same message with server error fails",
			err: &owlvigil.APIError{
				StatusCode: http.StatusInternalServerError,
				Code:       "internal_error",
				Message:    "feature.audit_logs is not included in the Team plan",
			},
			wantStatus: "FAIL",
		},
		{
			name: "matching status with different code fails",
			err: &owlvigil.APIError{
				StatusCode: http.StatusPaymentRequired,
				Code:       "billing_error",
				Message:    "feature.audit_logs is not included in the Team plan",
			},
			wantStatus: "FAIL",
		},
		{
			name: "matching status and code with different message fails",
			err: &owlvigil.APIError{
				StatusCode: http.StatusPaymentRequired,
				Code:       "upgrade_required",
				Message:    "another feature is not included in the Team plan",
			},
			wantStatus: "FAIL",
		},
		{
			name:       "plain error with matching text fails",
			err:        errors.New("feature.audit_logs is not included in the Team plan"),
			wantStatus: "FAIL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &runner{}
			r.callSkipKnown("list workspace audit logs", "GET /v1/workspaces/:workspace_id/audit-logs", expected, func() error {
				return tt.err
			})
			if len(r.steps) != 1 {
				t.Fatalf("runner.callSkipKnown(%q) recorded %d steps, want 1", tt.name, len(r.steps))
			}
			if got := r.steps[0].Status; got != tt.wantStatus {
				t.Errorf("runner.callSkipKnown(%q) status = %q, want %q", tt.name, got, tt.wantStatus)
			}
		})
	}
}

func TestRunner_CallSkipKnownMatchesQuotaExceededCode(t *testing.T) {
	r := &runner{}
	r.callSkipKnown(
		"create gateway key",
		"POST /v1/gateway/keys",
		expectedQuotaExceeded("quota.gateway_keys limit exceeded"),
		func() error {
			return &owlvigil.APIError{
				StatusCode: http.StatusPaymentRequired,
				Code:       "quota_exceeded",
				Message:    "quota.gateway_keys limit exceeded for this workspace",
			}
		},
	)

	if len(r.steps) != 1 {
		t.Fatalf("runner.callSkipKnown(quota_exceeded) recorded %d steps, want 1", len(r.steps))
	}
	if got := r.steps[0].Status; got != "SKIP" {
		t.Errorf("runner.callSkipKnown(quota_exceeded) status = %q, want SKIP", got)
	}
}

func TestRunner_WebhookMutationsUseTemporaryEndpointEvents(t *testing.T) {
	t.Setenv("OWLVIGIL_SMOKE_WEBHOOK_URL", "https://example.com/sdk-smoke-webhook")

	const (
		workspaceID      int64 = 1
		endpointID       int64 = 101
		temporaryEventID       = "202"
		otherEventID           = "999"
	)
	var retryEventID string
	var redeliverEventID string
	var bulkRequest management.BulkRedeliverRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/webhook-endpoints":
			_, _ = w.Write([]byte(`{"items":[],"page_info":{}}`))
		case req.Method == http.MethodPost && req.URL.Path == "/webhook-endpoints":
			_, _ = w.Write([]byte(`{"id":101}`))
		case req.Method == http.MethodGet && req.URL.Path == "/webhook-endpoints/101":
			_, _ = w.Write([]byte(`{"id":101}`))
		case req.Method == http.MethodPatch && req.URL.Path == "/webhook-endpoints/101":
			_, _ = w.Write([]byte(`{"id":101}`))
		case req.Method == http.MethodPost && (req.URL.Path == "/webhook-endpoints/101/enable" || req.URL.Path == "/webhook-endpoints/101/disable" || req.URL.Path == "/webhook-endpoints/101/rotate-secret"):
			_, _ = w.Write([]byte(`{"id":101}`))
		case req.Method == http.MethodPost && req.URL.Path == "/webhook-endpoints/101/test":
			_, _ = w.Write([]byte(`{"id":"202","endpoint_id":101,"status":"failed","attempts":1}`))
		case req.Method == http.MethodGet && req.URL.Path == "/webhook-endpoints/101/events":
			_, _ = w.Write([]byte(`{"items":[{"id":"202","endpoint_id":101,"status":"failed","attempts":1}],"page_info":{}}`))
		case req.Method == http.MethodGet && req.URL.Path == "/webhook-event-types":
			_, _ = w.Write([]byte(`[]`))
		case req.Method == http.MethodGet && req.URL.Path == "/webhook-events":
			_, _ = w.Write([]byte(`{"items":[{"id":"999","endpoint_id":999,"status":"failed","attempts":1},{"id":"202","endpoint_id":101,"status":"failed","attempts":1}],"page_info":{}}`))
		case req.Method == http.MethodGet && req.URL.Path == "/webhook-events/202":
			_, _ = w.Write([]byte(`{"id":"202","endpoint_id":101,"status":"failed","attempts":1}`))
		case req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, "/webhook-events/") && strings.HasSuffix(req.URL.Path, "/retry"):
			retryEventID = strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/webhook-events/"), "/retry")
			_, _ = w.Write([]byte(`{"id":"202","endpoint_id":101,"status":"pending","attempts":2}`))
		case req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, "/webhook-events/") && strings.HasSuffix(req.URL.Path, "/redeliver"):
			redeliverEventID = strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/webhook-events/"), "/redeliver")
			_, _ = w.Write([]byte(`{"id":"202","endpoint_id":101,"status":"pending","attempts":2}`))
		case req.Method == http.MethodPost && req.URL.Path == "/webhook-events/bulk-redeliver":
			if err := json.NewDecoder(req.Body).Decode(&bulkRequest); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"items":[{"id":"202","endpoint_id":101,"status":"pending"}],"page_info":{}}`))
		case req.Method == http.MethodDelete && req.URL.Path == "/webhook-endpoints/101":
			_, _ = w.Write([]byte(`{"message":"deleted"}`))
		default:
			http.Error(w, fmt.Sprintf("unexpected request: %s %s", req.Method, req.URL.Path), http.StatusNotFound)
		}
	}))
	defer server.Close()

	r := &runner{
		ctx: context.Background(),
		management: management.NewClient(
			owlvigil.WithBaseURL(server.URL),
			owlvigil.WithoutRetry(),
		),
		workspaceID: workspaceID,
		writes:      true,
	}
	r.runWebhooks()

	if retryEventID != temporaryEventID {
		t.Errorf("runWebhooks() retry event ID = %q, want temporary endpoint event %q; global first event was %q", retryEventID, temporaryEventID, otherEventID)
	}
	if redeliverEventID != temporaryEventID {
		t.Errorf("runWebhooks() redeliver event ID = %q, want temporary endpoint event %q; global first event was %q", redeliverEventID, temporaryEventID, otherEventID)
	}
	if bulkRequest.EndpointID == nil || *bulkRequest.EndpointID != endpointID {
		t.Errorf("runWebhooks() bulk endpoint ID = %v, want %d", bulkRequest.EndpointID, endpointID)
	}
	if got, want := bulkRequest.EventIDs, []int{202}; !slices.Equal(got, want) {
		t.Errorf("runWebhooks() bulk event IDs = %v, want %v", got, want)
	}
	for _, result := range r.steps {
		if result.Status == "FAIL" {
			t.Errorf("runWebhooks() step %q failed: %s", result.Name, result.Error)
		}
	}
}

func TestRunner_DoesNotCreateSmokeResourcesWhenWritesAreDisabled(t *testing.T) {
	var mutations atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			mutations.Add(1)
		}
		if req.URL.Path == "/webhook-event-types" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		if req.URL.Path == "/webhook-events" {
			_, _ = w.Write([]byte(`{"items":[{"id":"1"}],"page_info":{}}`))
			return
		}
		_, _ = w.Write([]byte(`{"items":[],"page_info":{}}`))
	}))
	defer server.Close()

	r := &runner{
		ctx: context.Background(),
		management: management.NewClient(
			owlvigil.WithBaseURL(server.URL),
			owlvigil.WithoutRetry(),
		),
		workspaceID: 1,
		writes:      false,
	}
	r.runGateway()
	r.runWebhooks()

	if got := mutations.Load(); got != 0 {
		t.Fatalf("mutating smoke requests = %d, want 0 when OWLVIGIL_SMOKE_WRITES is disabled", got)
	}
}

func TestRunner_CleanupContextOutlivesRunContext(t *testing.T) {
	runCtx, cancelRun := context.WithCancel(context.Background())
	cancelRun()
	r := &runner{ctx: runCtx}

	cleanupCtx, cancelCleanup := r.cleanupContext()
	defer cancelCleanup()

	if err := cleanupCtx.Err(); err != nil {
		t.Fatalf("cleanup context error = %v, want usable context", err)
	}
	deadline, ok := cleanupCtx.Deadline()
	if !ok {
		t.Fatal("cleanup context has no deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > cleanupTimeout {
		t.Fatalf("cleanup deadline remaining = %s, want (0, %s]", remaining, cleanupTimeout)
	}
}

func TestRunner_RequireAllRejectsSkippedUseCases(t *testing.T) {
	completeSteps := completeManagementSteps()
	skippedSteps := append([]step(nil), completeSteps...)
	skippedSteps[0].Status = "SKIP"

	tests := []struct {
		name       string
		requireAll bool
		steps      []step
		want       bool
	}{
		{
			name:       "strict complete",
			requireAll: true,
			steps:      completeSteps,
			want:       true,
		},
		{
			name:       "strict skip is incomplete",
			requireAll: true,
			steps:      skippedSteps,
			want:       false,
		},
		{
			name:       "OAuth skip does not affect Management completeness",
			requireAll: true,
			steps:      append(append([]step(nil), completeSteps...), step{Contract: "GET /oauth/authorize", Status: "SKIP"}),
			want:       true,
		},
		{
			name:       "duplicate cleanup skip does not erase a pass",
			requireAll: true,
			steps: append(
				append([]step(nil), completeSteps...),
				step{Contract: completeSteps[0].Contract, Status: "SKIP"},
			),
			want: true,
		},
		{
			name:  "non-strict skip is allowed but reported",
			steps: []step{{Status: "PASS"}, {Status: "SKIP"}},
			want:  true,
		},
		{
			name:       "failure is always incomplete",
			requireAll: true,
			steps:      []step{{Status: "FAIL"}},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &runner{requireAll: tt.requireAll, steps: tt.steps}
			if got := r.complete(); got != tt.want {
				t.Errorf("complete() = %t, want %t", got, tt.want)
			}
		})
	}
}

func completeManagementSteps() []step {
	steps := make([]step, managementOperationCount)
	for index := range steps {
		steps[index] = step{
			Contract: fmt.Sprintf("GET /v1/contract/%d", index),
			Status:   "PASS",
		}
	}
	return steps
}

func TestRunner_ConfiguredWrite(t *testing.T) {
	t.Run("writes disabled", func(t *testing.T) {
		t.Setenv("OWLVIGIL_TEST_CONFIGURED_WRITE", "configured")
		called := false
		r := &runner{writes: false}
		r.configuredWrite(
			"configured write",
			"POST /v1/configured",
			[]string{"OWLVIGIL_TEST_CONFIGURED_WRITE"},
			func([]string) error {
				called = true
				return nil
			},
		)
		if called {
			t.Fatal("configured write callback was called while writes were disabled")
		}
		if len(r.steps) != 1 || r.steps[0].Status != "SKIP" {
			t.Fatalf("steps = %+v, want one SKIP", r.steps)
		}
	})

	t.Run("missing environment", func(t *testing.T) {
		t.Setenv("OWLVIGIL_TEST_CONFIGURED_WRITE", "")
		called := false
		r := &runner{writes: true}
		r.configuredWrite(
			"configured write",
			"POST /v1/configured",
			[]string{"OWLVIGIL_TEST_CONFIGURED_WRITE"},
			func([]string) error {
				called = true
				return nil
			},
		)
		if called {
			t.Fatal("configured write callback was called without required environment")
		}
		if len(r.steps) != 1 || r.steps[0].Status != "SKIP" ||
			!strings.Contains(r.steps[0].Error, "OWLVIGIL_TEST_CONFIGURED_WRITE") {
			t.Fatalf("steps = %+v, want one descriptive SKIP", r.steps)
		}
	})

	t.Run("configured", func(t *testing.T) {
		t.Setenv("OWLVIGIL_TEST_CONFIGURED_WRITE", "configured")
		var got string
		r := &runner{writes: true}
		r.configuredWrite(
			"configured write",
			"POST /v1/configured",
			[]string{"OWLVIGIL_TEST_CONFIGURED_WRITE"},
			func(values []string) error {
				got = values[0]
				return nil
			},
		)
		if got != "configured" {
			t.Fatalf("configured value = %q, want configured", got)
		}
		if len(r.steps) != 1 || r.steps[0].Status != "PASS" {
			t.Fatalf("steps = %+v, want one PASS", r.steps)
		}
	})
}

func TestRunner_ConfiguredCallDoesNotRequireWrites(t *testing.T) {
	t.Setenv("OWLVIGIL_TEST_CONFIGURED_CALL", "configured")
	var got string
	r := &runner{writes: false}
	r.configuredCall(
		"configured call",
		"GET /v1/configured",
		[]string{"OWLVIGIL_TEST_CONFIGURED_CALL"},
		func(values []string) error {
			got = values[0]
			return nil
		},
	)
	if got != "configured" {
		t.Fatalf("configured value = %q, want configured", got)
	}
	if len(r.steps) != 1 || r.steps[0].Status != "PASS" {
		t.Fatalf("steps = %+v, want one PASS", r.steps)
	}
}
