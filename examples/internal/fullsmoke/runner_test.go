package fullsmoke

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
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
		t.Errorf("username = %v, want %q", got, profile.Username)
	}
	if got := requestBody["username"]; got == profile.Name {
		t.Errorf("username = display name %q, want account username", profile.Name)
	}
}

func TestRunnerCallSkipExpectedRequiresExactAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		code       string
		wantStatus string
	}{
		{
			name:       "documented feature gate is skipped",
			statusCode: http.StatusPaymentRequired,
			code:       "upgrade_required",
			wantStatus: "SKIP",
		},
		{
			name:       "server error with matching message fails",
			statusCode: http.StatusInternalServerError,
			code:       "upstream_error",
			wantStatus: "FAIL",
		},
		{
			name:       "unexpected code with matching status and message fails",
			statusCode: http.StatusPaymentRequired,
			code:       "invalid_request",
			wantStatus: "FAIL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &runner{}
			r.callSkipExpected(
				"list workspace audit logs",
				"GET /v1/workspaces/:workspace_id/audit-logs",
				auditLogsFeatureGate,
				func() error {
					return &owlvigil.APIError{
						StatusCode: tt.statusCode,
						Code:       tt.code,
						Message:    "feature.audit_logs is not included because entitlement lookup failed",
					}
				},
			)

			if len(r.steps) != 1 || r.steps[0].Status != tt.wantStatus {
				t.Fatalf("steps = %+v, want one %s", r.steps, tt.wantStatus)
			}
		})
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
