package main

import (
	"context"
	"fmt"
	"log"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/examples/internal/envfile"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	apiKey, err := envfile.Required("OWLVIGIL_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Initialize client with a service-account API key that has billing scopes.
	client := management.NewClient(
		owlvigil.WithAPIKey(apiKey),
	)

	// Example 1: List subscription plans
	fmt.Println("=== Subscription Plans ===")
	plans, meta, err := client.ListPlans(ctx, management.ListOptions{Limit: 10})
	if err != nil {
		log.Fatalf("Failed to list plans: %v", err)
	}
	fmt.Printf("Request ID: %s\n", meta.RequestID)
	for _, plan := range plans.Items {
		fmt.Printf("- %s: $%.2f/%s\n", plan.Name, plan.Price, plan.Interval)
	}

	// Example 2: Get current subscription
	fmt.Println("\n=== Current Subscription ===")
	subscription, _, err := client.GetSubscription(ctx)
	if err != nil {
		log.Printf("No active subscription: %v", err)
	} else {
		fmt.Printf("Plan: %s, Status: %s\n", subscription.PlanID, subscription.Status)
	}

	// Example 3: List payment methods
	fmt.Println("\n=== Payment Methods ===")
	methods, _, err := client.ListPaymentMethods(ctx, management.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list payment methods: %v", err)
	}
	for _, method := range methods.Items {
		defaultStr := ""
		if method.IsDefault {
			defaultStr = " (default)"
		}
		fmt.Printf("- %s **** %s%s\n", method.Brand, method.Last4, defaultStr)
	}

	// Example 4: Check billing balance
	fmt.Println("\n=== Balance ===")
	balance, _, err := client.GetBalance(ctx)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}
	fmt.Printf("Balance: %.2f %s\n", balance.Amount, balance.Currency)
}
