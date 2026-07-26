package management_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

// TestAllExecutableManagementUseCasesPassRefactoredOpenAPIFacade is opt-in so
// ordinary SDK consumers do not need a sibling OpenAPI checkout. The
// cross-repository alignment gate enables it. Unlike an SDK-only mock, this
// test passes every request through the real typed facade, including strict
// path, query, unknown-field, body-mode, and request DTO validation.
func TestAllExecutableManagementUseCasesPassRefactoredOpenAPIFacade(t *testing.T) {
	openAPIRepo := strings.TrimSpace(os.Getenv("OWLVIGIL_OPENAPI_REPO"))
	if openAPIRepo == "" {
		t.Skip("set OWLVIGIL_OPENAPI_REPO to run the cross-repository facade contract")
	}

	type upstreamRequest struct {
		method string
		path   string
		query  string
		body   []byte
	}
	upstreamRequests := make(chan upstreamRequest, 141)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read fake Dashboard request body: %v", err)
		}
		upstreamRequests <- upstreamRequest{
			method: request.Method,
			path:   request.URL.Path,
			query:  request.URL.RawQuery,
			body:   body,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"request_id": "req_fake_dashboard",
			"code":       "ok",
			"message":    "ok",
			"data":       nil,
		})
	}))
	defer upstream.Close()

	binary := filepath.Join(t.TempDir(), "owlvigil-openapi")
	build := exec.Command("go", "build", "-o", binary, "./cmd/openapi")
	build.Dir = openAPIRepo
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build current OpenAPI facade: %v\n%s", err, output)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve OpenAPI facade address: %v", err)
	}
	listenAddress := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("release OpenAPI facade address: %v", err)
	}

	var processOutput bytes.Buffer
	process := exec.Command(binary)
	process.Env = append(os.Environ(),
		"OPENAPI_LISTEN_ADDR="+listenAddress,
		"OPENAPI_DASHBOARD_URL="+upstream.URL,
		"OPENAPI_DASHBOARD_MANAGEMENT_HOST=api.owlvigil.test",
		"OPENAPI_REQUEST_TIMEOUT=5s",
		"OPENAPI_STARTUP_TIMEOUT=5s",
		"OPENAPI_SHUTDOWN_TIMEOUT=5s",
	)
	process.Stdout = &processOutput
	process.Stderr = &processOutput
	if err := process.Start(); err != nil {
		t.Fatalf("start current OpenAPI facade: %v", err)
	}
	processDone := make(chan error, 1)
	go func() {
		processDone <- process.Wait()
	}()
	t.Cleanup(func() {
		if process.Process == nil {
			return
		}
		_ = process.Process.Signal(os.Interrupt)
		select {
		case <-processDone:
		case <-time.After(5 * time.Second):
			_ = process.Process.Kill()
			<-processDone
		}
	})

	baseURL := "http://" + listenAddress
	if err := waitForOpenAPIHealth(t.Context(), baseURL+"/healthz", processDone); err != nil {
		t.Fatalf("wait for current OpenAPI facade: %v\n%s", err, processOutput.String())
	}

	client := management.NewClient(
		owlvigil.WithBaseURL(baseURL+"/v1"),
		owlvigil.WithAPIKey("management_contract_key"),
		owlvigil.WithoutRetry(),
	)
	expectedClient, expectedRequests := newManagementContractClient(t, `null`)
	for _, useCase := range allExecutableManagementUseCases() {
		if err := useCase.call(t.Context(), expectedClient); err != nil {
			t.Fatalf("%s baseline SDK call error = %v", useCase.contract, err)
		}
		expected := <-expectedRequests

		if err := useCase.call(t.Context(), client); err != nil {
			t.Errorf("%s rejected by current OpenAPI facade: %v", useCase.contract, err)
			continue
		}

		method, pattern, ok := strings.Cut(useCase.contract, " ")
		if !ok {
			t.Fatalf("invalid Management operation %q", useCase.contract)
		}
		select {
		case got := <-upstreamRequests:
			if got.method != method {
				t.Errorf("%s Dashboard upstream method = %q, want %q", useCase.contract, got.method, method)
			}
			wantPath := "/internal/openapi/v1" + executableManagementPath(pattern)
			if got.path != wantPath {
				t.Errorf("%s Dashboard upstream path = %q, want %q", useCase.contract, got.path, wantPath)
			}
			if got.query != expected.query.Encode() {
				t.Errorf(
					"%s Dashboard upstream query = %q, want %q",
					useCase.contract,
					got.query,
					expected.query.Encode(),
				)
			}
			assertForwardedManagementBody(t, useCase.contract, got.body, expected.body)
		case <-time.After(time.Second):
			t.Fatalf("%s did not reach the fake Dashboard upstream", useCase.contract)
		}
	}
}

func assertForwardedManagementBody(t *testing.T, contract string, got, want []byte) {
	t.Helper()
	got = bytes.TrimSpace(got)
	want = bytes.TrimSpace(want)
	if len(want) == 0 {
		if len(got) != 0 {
			t.Errorf("%s Dashboard upstream body = %s, want empty body", contract, got)
		}
		return
	}
	assertJSONSemanticallyEqual(t, got, string(want))
}

func waitForOpenAPIHealth(ctx context.Context, healthURL string, processDone <-chan error) error {
	deadline := time.NewTimer(20 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	client := &http.Client{Timeout: 250 * time.Millisecond}

	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err != nil {
			return err
		}
		response, err := client.Do(request)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}

		select {
		case err := <-processDone:
			if err == nil {
				return context.Canceled
			}
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return context.DeadlineExceeded
		case <-ticker.C:
		}
	}
}
