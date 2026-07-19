# Billing & Subscription Guide

> This guide covers payment workflows. For budgets, spend limits, thresholds,
> quota, and request-level spend evidence, see
> [Financial governance](financial-governance.md).

Complete guide to managing billing, subscriptions, and payments with the OwlVigil Go SDK.

---

## Overview

The Billing API allows you to:
- 📋 List and view subscription plans
- 💳 Create and manage subscriptions
- 💰 Handle top-ups and credits
- 🔐 Manage payment methods
- 📄 View invoices and orders

---

## Authentication

All billing operations require a service-account API key with the appropriate billing scopes:

```go
client := management.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")))
```

---

## Subscription Plans

### List Available Plans

```go
plans, meta, err := client.ListPlans(ctx, management.ListOptions{})
if err != nil {
    log.Fatalf("Failed to list plans: %v", err)
}

for _, plan := range plans.Items {
    fmt.Printf("Plan: %s - $%.2f/%s\n", plan.Name, plan.Price, plan.Interval)
    fmt.Printf("  Features: %v\n", plan.Features)
}
```

### Get Plan Details

```go
plan, _, err := client.GetPlan(ctx, "pro")
if err != nil {
    log.Fatalf("Failed to get plan: %v", err)
}

fmt.Printf("Plan: %s\n", plan.Name)
fmt.Printf("Price: $%.2f/%s\n", plan.Price, plan.Interval)
```

---

## Subscription Management

### Get Current Subscription

```go
subscription, _, err := client.GetSubscription(ctx)
if err != nil {
    log.Printf("No active subscription: %v", err)
} else {
    fmt.Printf("Plan: %s\n", subscription.PlanID)
    fmt.Printf("Status: %s\n", subscription.Status)
    fmt.Printf("Period End: %s\n", subscription.CurrentPeriodEnd)
}
```

### Create Subscription (Checkout Flow)

```go
// Redirect user to Stripe Checkout
checkout, _, err := client.CreateSubscriptionCheckout(ctx, &management.CreateSubscriptionCheckoutRequest{
    PlanID:     "pro",
    Interval:   "monthly",
    SuccessURL: "https://yourapp.com/success",
    CancelURL:  "https://yourapp.com/cancel",
})
if err != nil {
    log.Fatalf("Failed to create checkout: %v", err)
}

// Redirect user to checkout.CheckoutURL
fmt.Printf("Redirect to: %s\n", checkout.CheckoutURL)

// After payment, check session status
session, _, err := client.GetSubscriptionCheckoutSession(ctx, checkout.SessionID)
if err != nil {
    log.Fatalf("Failed to get session: %v", err)
}
fmt.Printf("Session status: %s\n", session.Status)
```

### Create Subscription (In-App Flow)

```go
// Step 1: Create payment intent
intent, _, err := client.CreateSubscriptionInApp(ctx, &management.CreateSubscriptionInAppRequest{
    PlanID:   "pro",
    Interval: "monthly",
})
if err != nil {
    log.Fatalf("Failed to create payment: %v", err)
}

// Step 2: Client confirms payment with Stripe.js
// (Use intent.ClientSecret on frontend)

// Step 3: Confirm subscription after 3DS
subscription, _, err := client.ConfirmSubscriptionInApp(ctx, &management.ConfirmSubscriptionInAppRequest{
    PaymentIntentID: intent.PaymentIntentID,
})
if err != nil {
    log.Fatalf("Failed to confirm: %v", err)
}
fmt.Printf("Subscription created: %s\n", subscription.ID)
```

### Upgrade Subscription

```go
subscription, _, err := client.UpgradeSubscription(ctx, &management.UpgradeSubscriptionRequest{
    NewPlanID: "enterprise",
})
if err != nil {
    log.Fatalf("Failed to upgrade: %v", err)
}
fmt.Printf("Upgraded to: %s\n", subscription.PlanID)
```

### Downgrade Subscription

```go
// Downgrade is scheduled for end of period
subscription, _, err := client.DowngradeSubscription(ctx, &management.DowngradeSubscriptionRequest{
    NewPlanID: "starter",
})
if err != nil {
    log.Fatalf("Failed to downgrade: %v", err)
}
fmt.Printf("Will downgrade to: %s\n", *subscription.PendingDowngrade)
```

### Cancel Subscription

```go
// Cancellation takes effect at period end
subscription, _, err := client.CancelSubscription(ctx)
if err != nil {
    log.Fatalf("Failed to cancel: %v", err)
}

if subscription.CancelAtPeriodEnd {
    fmt.Printf("Subscription will cancel on: %s\n", subscription.CurrentPeriodEnd)
}
```

### Reactivate Subscription

```go
subscription, _, err := client.ReactivateSubscription(ctx)
if err != nil {
    log.Fatalf("Failed to reactivate: %v", err)
}
fmt.Printf("Subscription reactivated: %s\n", subscription.Status)
```

---

## Top-ups

### List Top-up Plans

```go
plans, _, err := client.ListTopupPlans(ctx, management.ListOptions{})
if err != nil {
    log.Fatalf("Failed to list top-up plans: %v", err)
}

for _, plan := range plans.Items {
    bonus := ""
    if plan.BonusAmount > 0 {
        bonus = fmt.Sprintf(" + $%.2f bonus", plan.BonusAmount)
    }
    fmt.Printf("Top-up: $%.2f%s\n", plan.Amount, bonus)
}
```

### Create Top-up (Checkout Flow)

```go
checkout, _, err := client.CreateTopupCheckout(ctx, &management.CreateTopupCheckoutRequest{
    Amount:     100.0,
    SuccessURL: "https://yourapp.com/success",
    CancelURL:  "https://yourapp.com/cancel",
})
if err != nil {
    log.Fatalf("Failed to create checkout: %v", err)
}

// Redirect user to checkout.CheckoutURL
fmt.Printf("Redirect to: %s\n", checkout.CheckoutURL)
```

### Create Top-up (In-App Flow)

```go
// Step 1: Create payment intent
intent, _, err := client.CreateTopupInApp(ctx, &management.CreateTopupInAppRequest{
    Amount: 100.0,
})
if err != nil {
    log.Fatalf("Failed to create payment: %v", err)
}

// Step 2: Client confirms payment with Stripe.js

// Step 3: Confirm top-up
topup, _, err := client.ConfirmTopupInApp(ctx, &management.ConfirmTopupInAppRequest{
    PaymentIntentID: intent.PaymentIntentID,
})
if err != nil {
    log.Fatalf("Failed to confirm: %v", err)
}
fmt.Printf("Top-up completed: $%.2f\n", topup.Amount)
```

### List Top-up History

```go
topups, _, err := client.ListTopups(ctx, management.ListOptions{Limit: 10})
if err != nil {
    log.Fatalf("Failed to list top-ups: %v", err)
}

for _, topup := range topups.Items {
    fmt.Printf("%s: $%.2f - %s\n", topup.CreatedAt, topup.Amount, topup.Status)
}
```

---

## Payment Methods

### List Payment Methods

```go
methods, _, err := client.ListPaymentMethods(ctx, management.ListOptions{})
if err != nil {
    log.Fatalf("Failed to list payment methods: %v", err)
}

for _, method := range methods.Items {
    defaultStr := ""
    if method.IsDefault {
        defaultStr = " (default)"
    }
    fmt.Printf("%s **** %s - Exp: %d/%d%s\n",
        method.Brand, method.Last4, method.ExpMonth, method.ExpYear, defaultStr)
}
```

### Add Payment Method

```go
// Step 1: Create SetupIntent
setupIntent, _, err := client.CreatePaymentMethodSetupIntent(ctx)
if err != nil {
    log.Fatalf("Failed to create setup intent: %v", err)
}

// Step 2: Client confirms with Stripe.js
// (Use setupIntent.ClientSecret on frontend)

// Step 3: Save payment method
paymentMethod, _, err := client.SavePaymentMethod(ctx, &management.SavePaymentMethodRequest{
    PaymentMethodID: "pm_xxxxx", // from Stripe.js
    SetAsDefault:    true,
})
if err != nil {
    log.Fatalf("Failed to save payment method: %v", err)
}
fmt.Printf("Payment method saved: %s\n", paymentMethod.ID)
```

### Set Default Payment Method

```go
_, _, err := client.SetDefaultPaymentMethod(ctx, "pm_xxxxx")
if err != nil {
    log.Fatalf("Failed to set default: %v", err)
}
fmt.Println("Default payment method updated")
```

### Delete Payment Method

```go
_, err := client.DeletePaymentMethod(ctx, "pm_xxxxx")
if err != nil {
    log.Fatalf("Failed to delete payment method: %v", err)
}
fmt.Println("Payment method deleted")
```

---

## Billing Information

### Get Billing Overview

```go
overview, _, err := client.GetBillingOverview(ctx)
if err != nil {
    log.Fatalf("Failed to get overview: %v", err)
}

if overview.Balance != nil {
    fmt.Printf("Balance: $%.2f %s\n", overview.Balance.Amount, overview.Balance.Currency)
}

if overview.Subscription != nil {
    fmt.Printf("Plan: %s (%s)\n", overview.Subscription.PlanID, overview.Subscription.Status)
}
```

### Get/Update Billing Details

```go
// Get current billing details
details, _, err := client.GetBillingDetails(ctx, workspaceID)
if err != nil {
    log.Fatalf("Failed to get billing details: %v", err)
}

// Update billing details
companyName := "Acme Corporation"
email := "billing@acme.com"
updated, _, err := client.UpdateBillingDetails(ctx, workspaceID, &management.UpdateBillingDetailsRequest{
    CompanyName: &companyName,
    Email:       &email,
    Address: &management.Address{
        Line1:      "123 Main St",
        City:       "San Francisco",
        State:      "CA",
        PostalCode: "94105",
        Country:    "US",
    },
})
if err != nil {
    log.Fatalf("Failed to update: %v", err)
}
fmt.Printf("Billing details updated for: %s\n", updated.CompanyName)
```

### List Invoices

```go
invoices, _, err := client.ListInvoices(ctx, management.ListOptions{Limit: 10})
if err != nil {
    log.Fatalf("Failed to list invoices: %v", err)
}

for _, invoice := range invoices.Items {
    fmt.Printf("Invoice %s: $%.2f - %s\n", invoice.ID, invoice.Amount, invoice.Status)
}
```

### Get Invoice Details

```go
invoice, _, err := client.GetInvoice(ctx, "inv_xxxxx")
if err != nil {
    log.Fatalf("Failed to get invoice: %v", err)
}

fmt.Printf("Invoice: %s\n", invoice.ID)
fmt.Printf("Amount: $%.2f %s\n", invoice.Amount, invoice.Currency)
fmt.Printf("Status: %s\n", invoice.Status)
if invoice.DueDate != "" {
    fmt.Printf("Due: %s\n", invoice.DueDate)
}
```

---

## Orders

### List Orders

```go
orders, _, err := client.ListOrders(ctx, management.ListOptions{}, "subscription")
if err != nil {
    log.Fatalf("Failed to list orders: %v", err)
}

for _, order := range orders.Items {
    fmt.Printf("Order %s: %s - $%.2f (%s)\n",
        order.ID, order.Type, order.Amount, order.Status)
}
```

### Get Order Details

```go
order, _, err := client.GetOrder(ctx, "order_xxxxx")
if err != nil {
    log.Fatalf("Failed to get order: %v", err)
}

fmt.Printf("Order: %s\n", order.ID)
fmt.Printf("Type: %s\n", order.Type)
fmt.Printf("Amount: $%.2f\n", order.Amount)
fmt.Printf("Status: %s\n", order.Status)
```

---

## Error Handling

### Common Errors

```go
subscription, _, err := client.GetSubscription(ctx)
if err != nil {
    // Check for specific error types
    if strings.Contains(err.Error(), "not found") {
        fmt.Println("No active subscription")
    } else if strings.Contains(err.Error(), "not yet implemented") {
        fmt.Println("This feature is being rolled out")
    } else {
        log.Fatalf("Unexpected error: %v", err)
    }
    return
}
```

---

## Best Practices

### 1. Handle Webhook Events

Always verify webhook signatures and handle subscription events:

```go
// See webhooks.md for webhook verification
```

### 2. Check Balance Before Operations

```go
balance, _, err := client.GetBalance(ctx)
if err != nil {
    return err
}

if balance.Amount < 10 {
    // Prompt user to top up
    fmt.Println("Low balance, please add credits")
}
```

### 3. Graceful Subscription Status Handling

```go
subscription, _, err := client.GetSubscription(ctx)
if err != nil {
    // User might not have a subscription
    return nil
}

switch subscription.Status {
case "active":
    // Normal operation
case "past_due":
    // Payment failed, prompt to update payment method
case "canceled":
    // Subscription ended
case "trialing":
    // User is in trial period
}
```

### 4. Use Idempotency for Payments

```go
import "github.com/google/uuid"

checkout, _, err := client.CreateSubscriptionCheckout(
    ctx,
    &management.CreateSubscriptionCheckoutRequest{
        PlanID: "pro",
        Interval: "monthly",
        SuccessURL: "https://app.com/success",
        CancelURL: "https://app.com/cancel",
    },
    owlvigil.WithIdempotencyKey(uuid.New().String()),
)
```

---

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    owlvigil "github.com/Syrovex/owlvigil_sdk_go"
    "github.com/Syrovex/owlvigil_sdk_go/management"
)

func main() {
    ctx := context.Background()
    client := management.NewClient(
        owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
    )

    // List available plans
    plans, _, err := client.ListPlans(ctx, management.ListOptions{})
    if err != nil {
        log.Fatalf("Failed to list plans: %v", err)
    }

    fmt.Println("Available Plans:")
    for _, plan := range plans.Items {
        fmt.Printf("  %s: $%.2f/%s\n", plan.Name, plan.Price, plan.Interval)
    }

    // Check current subscription
    subscription, _, err := client.GetSubscription(ctx)
    if err != nil {
        fmt.Println("No active subscription")
    } else {
        fmt.Printf("\nCurrent Plan: %s (%s)\n", subscription.PlanID, subscription.Status)
    }

    // Check balance
    balance, _, err := client.GetBalance(ctx)
    if err != nil {
        log.Fatalf("Failed to get balance: %v", err)
    }
    fmt.Printf("\nBalance: $%.2f %s\n", balance.Amount, balance.Currency)

    // List payment methods
    methods, _, err := client.ListPaymentMethods(ctx, management.ListOptions{})
    if err != nil {
        log.Fatalf("Failed to list payment methods: %v", err)
    }

    fmt.Println("\nPayment Methods:")
    for _, method := range methods.Items {
        defaultStr := ""
        if method.IsDefault {
            defaultStr = " (default)"
        }
        fmt.Printf("  %s **** %s%s\n", method.Brand, method.Last4, defaultStr)
    }
}
```

---

## See Also

- [Authentication Guide](./authentication.md)
- [Error Handling](./errors.md)
- [Webhooks](./webhooks.md)
- [Examples](../examples/billing-subscription/)
