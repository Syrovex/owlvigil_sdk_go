package management_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

const autoChargeResponse = `{
	"configured": true,
	"workspace_id": 17,
	"workspace_name": "Acme",
	"enabled": true,
	"threshold_amount": 1,
	"charge_amount": 50,
	"currency": "usd",
	"status": "active",
	"payer_user_id": 9,
	"payer_name": "Owner",
	"payment_method": {"brand":"visa","last4":"4242"}
}`

func TestClient_GetAutoCharge_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()
	client, requests := newManagementContractClient(t, autoChargeResponse)

	got, _, err := client.GetAutoCharge(t.Context(), 17)
	if err != nil {
		t.Fatalf("GetAutoCharge() error = %v, want nil", err)
	}
	if got.WorkspaceID != 17 || got.ChargeAmount != 50 || got.PaymentMethod == nil || got.PaymentMethod.Last4 != "4242" {
		t.Errorf("GetAutoCharge() = %+v, want complete auto-charge fields", got)
	}
	assertManagementRequest(t, requests, http.MethodGet, "/billing/auto-charge", url.Values{"workspace_id": {"17"}}, "")
}

func TestClient_UpdateAutoCharge_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()
	client, requests := newManagementContractClient(t, autoChargeResponse)

	_, _, err := client.UpdateAutoCharge(t.Context(), &management.UpdateAutoChargeRequest{
		WorkspaceID: 17, Enabled: true, ChargeAmount: 50, TransferPayer: true,
	})
	if err != nil {
		t.Fatalf("UpdateAutoCharge() error = %v, want nil", err)
	}
	assertManagementRequest(t, requests, http.MethodPut, "/billing/auto-charge", url.Values{}, `{"workspace_id":17,"enabled":true,"charge_amount":50,"transfer_payer":true}`)
}

func TestClient_RetryAutoCharge_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()
	client, requests := newManagementContractClient(t, autoChargeResponse)

	_, _, err := client.RetryAutoCharge(t.Context(), 17)
	if err != nil {
		t.Fatalf("RetryAutoCharge() error = %v, want nil", err)
	}
	assertManagementRequest(t, requests, http.MethodPost, "/billing/auto-charge/retry", url.Values{}, `{"workspace_id":17}`)
}

func TestTopupAndOrder_DecodeAutoChargeReceiptFields(t *testing.T) {
	t.Parallel()
	client, topupRequests := newManagementContractClient(t, `{"id":23,"amount":50,"currency":"usd","status":"succeeded","source":"auto_charge","receipt_url":"https://receipt.example/23"}`)

	topup, _, err := client.GetTopup(t.Context(), "23")
	if err != nil {
		t.Fatalf("GetTopup() error = %v, want nil", err)
	}
	if topup.Source != "auto_charge" || topup.ReceiptURL == nil || *topup.ReceiptURL != "https://receipt.example/23" {
		t.Errorf("GetTopup() = %+v, want auto-charge receipt fields", topup)
	}
	assertManagementRequest(t, topupRequests, http.MethodGet, "/billing/topups/23", url.Values{}, "")

	legacyClient, orderRequests := newManagementContractClient(t, `{"id":24,"amount":5,"currency":"usd","status":"succeeded"}`)
	order, _, err := legacyClient.GetOrder(t.Context(), "24")
	if err != nil {
		t.Fatalf("GetOrder() legacy response error = %v, want nil", err)
	}
	if order.Source != "" || order.ReceiptURL != nil {
		t.Errorf("GetOrder() legacy response = %+v, want empty additive fields", order)
	}
	assertManagementRequest(t, orderRequests, http.MethodGet, "/billing/orders/24", url.Values{}, "")
}

func TestClient_AutoChargeMethods_ReturnAPIError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"invalid_auto_charge","message":"invalid automatic charge"}`))
	}))
	t.Cleanup(server.Close)
	client := management.NewClient(
		owlvigil.WithBaseURL(server.URL),
		owlvigil.WithAPIKey("management_test_key"),
		owlvigil.WithoutRetry(),
	)

	calls := []struct {
		name string
		call func() error
	}{
		{name: "get", call: func() error { _, _, err := client.GetAutoCharge(t.Context(), 17); return err }},
		{name: "update", call: func() error {
			_, _, err := client.UpdateAutoCharge(t.Context(), &management.UpdateAutoChargeRequest{WorkspaceID: 17, Enabled: true, ChargeAmount: 50})
			return err
		}},
		{name: "retry", call: func() error { _, _, err := client.RetryAutoCharge(t.Context(), 17); return err }},
	}
	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			var apiErr *owlvigil.APIError
			if err := tt.call(); !errors.As(err, &apiErr) {
				t.Fatalf("auto-charge call error = %T %v, want *owlvigil.APIError", err, err)
			}
			if apiErr.StatusCode != http.StatusBadRequest || apiErr.Code != "invalid_auto_charge" {
				t.Errorf("API error = %+v, want status 400 and code invalid_auto_charge", apiErr)
			}
		})
	}
}
