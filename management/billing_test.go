package management_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/management"
)

func TestSubscriptionEndpoints(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /billing/plans":                                       `{"items":[{"id":"starter","name":"Starter","price":10,"currency":"USD","interval":"monthly"}],"page_info":{}}`,
		"GET /billing/plans/starter":                               `{"id":"starter","name":"Starter","price":10,"currency":"USD","interval":"monthly"}`,
		"GET /billing/subscription":                                `{"id":"sub_1","plan_id":"starter","status":"active","current_period_end":"2026-08-01","cancel_at_period_end":false}`,
		"POST /billing/subscription/checkout":                      `{"checkout_url":"https://checkout.stripe.com/session_1","session_id":"cs_1","order_id":"order_1"}`,
		"POST /billing/subscription/in-app":                        `{"payment_intent_id":"pi_1","client_secret":"secret_1","requires_action":false,"status":"succeeded"}`,
		"POST /billing/subscription/in-app/confirm":                `{"id":"sub_1","plan_id":"starter","status":"active"}`,
		"POST /billing/subscription/upgrade":                       `{"id":"sub_1","plan_id":"pro","status":"active"}`,
		"POST /billing/subscription/downgrade":                     `{"id":"sub_1","plan_id":"starter","status":"active","pending_downgrade":"free"}`,
		"POST /billing/subscription/cancel":                        `{"id":"sub_1","plan_id":"starter","status":"active","cancel_at_period_end":true}`,
		"POST /billing/subscription/reactivate":                    `{"id":"sub_1","plan_id":"starter","status":"active","cancel_at_period_end":false}`,
		"GET /billing/subscription/checkout-sessions/cs_1":         `{"session_id":"cs_1","status":"complete","order_id":"order_1"}`,
		"POST /billing/subscription/checkout-sessions/sync-latest": `{"session_id":"cs_2","status":"complete","order_id":"order_2"}`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		body, ok := expected[key]
		if !ok {
			t.Fatalf("unexpected request %s", key)
		}
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := management.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAccessToken("test"))
	ctx := context.Background()

	// Test ListPlans
	plans, _, err := client.ListPlans(ctx, management.ListOptions{})
	if err != nil {
		t.Fatalf("ListPlans failed: %v", err)
	}
	if len(plans.Items) != 1 || plans.Items[0].ID != "starter" {
		t.Fatalf("plans = %+v", plans)
	}

	// Test GetPlan
	plan, _, err := client.GetPlan(ctx, "starter")
	if err != nil {
		t.Fatalf("GetPlan failed: %v", err)
	}
	if plan.ID != "starter" || plan.Price != 10 {
		t.Fatalf("plan = %+v", plan)
	}

	// Test GetSubscription
	sub, _, err := client.GetSubscription(ctx)
	if err != nil {
		t.Fatalf("GetSubscription failed: %v", err)
	}
	if sub.ID != "sub_1" || sub.Status != "active" {
		t.Fatalf("subscription = %+v", sub)
	}

	// Test CreateSubscriptionCheckout
	checkout, _, err := client.CreateSubscriptionCheckout(ctx, &management.CreateSubscriptionCheckoutRequest{
		PlanID:     "starter",
		Interval:   "monthly",
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
	})
	if err != nil {
		t.Fatalf("CreateSubscriptionCheckout failed: %v", err)
	}
	if checkout.SessionID != "cs_1" {
		t.Fatalf("checkout = %+v", checkout)
	}

	// Test CreateSubscriptionInApp
	inApp, _, err := client.CreateSubscriptionInApp(ctx, &management.CreateSubscriptionInAppRequest{
		PlanID:   "starter",
		Interval: "monthly",
	})
	if err != nil {
		t.Fatalf("CreateSubscriptionInApp failed: %v", err)
	}
	if inApp.PaymentIntentID != "pi_1" {
		t.Fatalf("inApp = %+v", inApp)
	}

	// Test ConfirmSubscriptionInApp
	_, _, err = client.ConfirmSubscriptionInApp(ctx, &management.ConfirmSubscriptionInAppRequest{
		PaymentIntentID: "pi_1",
	})
	if err != nil {
		t.Fatalf("ConfirmSubscriptionInApp failed: %v", err)
	}

	// Test UpgradeSubscription
	_, _, err = client.UpgradeSubscription(ctx, &management.UpgradeSubscriptionRequest{NewPlanID: "pro"})
	if err != nil {
		t.Fatalf("UpgradeSubscription failed: %v", err)
	}

	// Test DowngradeSubscription
	_, _, err = client.DowngradeSubscription(ctx, &management.DowngradeSubscriptionRequest{NewPlanID: "free"})
	if err != nil {
		t.Fatalf("DowngradeSubscription failed: %v", err)
	}

	// Test CancelSubscription
	_, _, err = client.CancelSubscription(ctx)
	if err != nil {
		t.Fatalf("CancelSubscription failed: %v", err)
	}

	// Test ReactivateSubscription
	_, _, err = client.ReactivateSubscription(ctx)
	if err != nil {
		t.Fatalf("ReactivateSubscription failed: %v", err)
	}

	// Test GetSubscriptionCheckoutSession
	session, _, err := client.GetSubscriptionCheckoutSession(ctx, "cs_1")
	if err != nil {
		t.Fatalf("GetSubscriptionCheckoutSession failed: %v", err)
	}
	if session.SessionID != "cs_1" {
		t.Fatalf("session = %+v", session)
	}

	// Test SyncLatestSubscriptionCheckout
	_, _, err = client.SyncLatestSubscriptionCheckout(ctx)
	if err != nil {
		t.Fatalf("SyncLatestSubscriptionCheckout failed: %v", err)
	}
}

func TestTopupEndpoints(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /billing/topup-plans":            `{"items":[{"id":"t_100","amount":100,"currency":"USD"}],"page_info":{}}`,
		"POST /billing/topups/checkout":       `{"checkout_url":"https://checkout.stripe.com/session_2","session_id":"cs_2","order_id":"order_2"}`,
		"POST /billing/topups/in-app":         `{"payment_intent_id":"pi_2","client_secret":"secret_2","requires_action":false,"status":"succeeded"}`,
		"POST /billing/topups/in-app/confirm": `{"id":"top_1","amount":100,"currency":"USD","status":"succeeded"}`,
		"GET /billing/topups":                 `{"items":[{"id":"top_1","amount":100,"currency":"USD","status":"succeeded"}],"page_info":{}}`,
		"GET /billing/topups/top_1":           `{"id":"top_1","amount":100,"currency":"USD","status":"succeeded"}`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		body, ok := expected[key]
		if !ok {
			t.Fatalf("unexpected request %s", key)
		}
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := management.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAccessToken("test"))
	ctx := context.Background()

	// Test ListTopupPlans
	plans, _, err := client.ListTopupPlans(ctx, management.ListOptions{})
	if err != nil {
		t.Fatalf("ListTopupPlans failed: %v", err)
	}
	if len(plans.Items) != 1 || plans.Items[0].Amount != 100 {
		t.Fatalf("plans = %+v", plans)
	}

	// Test CreateTopupCheckout
	checkout, _, err := client.CreateTopupCheckout(ctx, &management.CreateTopupCheckoutRequest{
		Amount:     100,
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
	})
	if err != nil {
		t.Fatalf("CreateTopupCheckout failed: %v", err)
	}
	if checkout.SessionID != "cs_2" {
		t.Fatalf("checkout = %+v", checkout)
	}

	// Test CreateTopupInApp
	_, _, err = client.CreateTopupInApp(ctx, &management.CreateTopupInAppRequest{Amount: 100})
	if err != nil {
		t.Fatalf("CreateTopupInApp failed: %v", err)
	}

	// Test ConfirmTopupInApp
	topup, _, err := client.ConfirmTopupInApp(ctx, &management.ConfirmTopupInAppRequest{PaymentIntentID: "pi_2"})
	if err != nil {
		t.Fatalf("ConfirmTopupInApp failed: %v", err)
	}
	if topup.Amount != 100 {
		t.Fatalf("topup = %+v", topup)
	}

	// Test ListTopups
	topups, _, err := client.ListTopups(ctx, management.ListOptions{})
	if err != nil {
		t.Fatalf("ListTopups failed: %v", err)
	}
	if len(topups.Items) != 1 {
		t.Fatalf("topups = %+v", topups)
	}

	// Test GetTopup
	_, _, err = client.GetTopup(ctx, "top_1")
	if err != nil {
		t.Fatalf("GetTopup failed: %v", err)
	}
}

func TestPaymentMethodsEndpoints(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /billing/payment-methods":               `{"items":[{"id":"pm_1","type":"card","brand":"visa","last4":"4242","is_default":true}],"page_info":{}}`,
		"POST /billing/payment-methods/setup-intent": `{"setup_intent_id":"seti_1","client_secret":"secret_1","status":"succeeded"}`,
		"POST /billing/payment-methods":              `{"id":"pm_2","type":"card","brand":"mastercard","last4":"5555","is_default":false}`,
		"PUT /billing/payment-methods/pm_2/default":  `{"id":"pm_2","type":"card","brand":"mastercard","last4":"5555","is_default":true}`,
		"DELETE /billing/payment-methods/pm_1":       `{}`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		body, ok := expected[key]
		if !ok {
			t.Fatalf("unexpected request %s", key)
		}
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := management.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAccessToken("test"))
	ctx := context.Background()

	// Test ListPaymentMethods
	methods, _, err := client.ListPaymentMethods(ctx, management.ListOptions{})
	if err != nil {
		t.Fatalf("ListPaymentMethods failed: %v", err)
	}
	if len(methods.Items) != 1 || methods.Items[0].Brand != "visa" {
		t.Fatalf("methods = %+v", methods)
	}

	// Test CreatePaymentMethodSetupIntent
	setupIntent, _, err := client.CreatePaymentMethodSetupIntent(ctx)
	if err != nil {
		t.Fatalf("CreatePaymentMethodSetupIntent failed: %v", err)
	}
	if setupIntent.SetupIntentID != "seti_1" {
		t.Fatalf("setupIntent = %+v", setupIntent)
	}

	// Test SavePaymentMethod
	pm, _, err := client.SavePaymentMethod(ctx, &management.SavePaymentMethodRequest{
		PaymentMethodID: "pm_2",
		SetAsDefault:    false,
	})
	if err != nil {
		t.Fatalf("SavePaymentMethod failed: %v", err)
	}
	if pm.ID != "pm_2" {
		t.Fatalf("pm = %+v", pm)
	}

	// Test SetDefaultPaymentMethod
	_, _, err = client.SetDefaultPaymentMethod(ctx, "pm_2")
	if err != nil {
		t.Fatalf("SetDefaultPaymentMethod failed: %v", err)
	}

	// Test DeletePaymentMethod
	_, err = client.DeletePaymentMethod(ctx, "pm_1")
	if err != nil {
		t.Fatalf("DeletePaymentMethod failed: %v", err)
	}
}

func TestBillingEndpoints(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /billing/overview":                               `{"balance":{"amount":100,"currency":"USD"},"subscription":{"id":"sub_1","plan_id":"starter","status":"active"}}`,
		"GET /workspaces/1/billing-details":                   `{"company_name":"Acme Inc","email":"billing@acme.com"}`,
		"PUT /workspaces/1/billing-details":                   `{"company_name":"Acme Corporation","email":"billing@acme.com"}`,
		"GET /billing/invoices/inv_1":                         `{"id":"inv_1","amount":100,"currency":"USD","status":"paid"}`,
		"GET /billing/orders":                                 `{"items":[{"id":"order_1","type":"subscription","amount":10,"status":"completed"}],"page_info":{}}`,
		"GET /billing/orders/order_1":                         `{"id":"order_1","type":"subscription","amount":10,"status":"completed"}`,
		"POST /billing/orders/order_1/confirm-stripe-session": `{"id":"order_1","type":"subscription","amount":10,"status":"completed"}`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		body, ok := expected[key]
		if !ok {
			t.Fatalf("unexpected request %s", key)
		}
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := management.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAccessToken("test"))
	ctx := context.Background()

	// Test GetBillingOverview
	overview, _, err := client.GetBillingOverview(ctx)
	if err != nil {
		t.Fatalf("GetBillingOverview failed: %v", err)
	}
	if overview.Balance == nil || overview.Balance.Amount != 100 {
		t.Fatalf("overview = %+v", overview)
	}

	// Test GetBillingDetails
	details, _, err := client.GetBillingDetails(ctx, 1)
	if err != nil {
		t.Fatalf("GetBillingDetails failed: %v", err)
	}
	if details.CompanyName != "Acme Inc" {
		t.Fatalf("details = %+v", details)
	}

	// Test UpdateBillingDetails
	companyName := "Acme Corporation"
	_, _, err = client.UpdateBillingDetails(ctx, 1, &management.UpdateBillingDetailsRequest{
		CompanyName: &companyName,
	})
	if err != nil {
		t.Fatalf("UpdateBillingDetails failed: %v", err)
	}

	// Test GetInvoice
	invoice, _, err := client.GetInvoice(ctx, "inv_1")
	if err != nil {
		t.Fatalf("GetInvoice failed: %v", err)
	}
	if invoice.ID != "inv_1" || invoice.Amount != 100 {
		t.Fatalf("invoice = %+v", invoice)
	}

	// Test ListOrders
	orders, _, err := client.ListOrders(ctx, management.ListOptions{}, "")
	if err != nil {
		t.Fatalf("ListOrders failed: %v", err)
	}
	if len(orders.Items) != 1 {
		t.Fatalf("orders = %+v", orders)
	}

	// Test GetOrder
	order, _, err := client.GetOrder(ctx, "order_1")
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if order.ID != "order_1" {
		t.Fatalf("order = %+v", order)
	}

	// Test ConfirmStripeSession
	_, _, err = client.ConfirmStripeSession(ctx, "order_1", &management.ConfirmStripeSessionRequest{
		SessionID: "cs_1",
	})
	if err != nil {
		t.Fatalf("ConfirmStripeSession failed: %v", err)
	}
}
