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
	workspaceID := int64(1) // Replace with your workspace ID

	// Initialize client with a service-account API key that has financial scopes.
	client := management.NewClient(
		owlvigil.WithAPIKey(apiKey),
	)

	// Example 1: Get financial governance config
	fmt.Println("=== Financial Governance ===")
	config, _, err := client.GetFinancialGovernance(ctx, workspaceID)
	if err != nil {
		log.Fatalf("Failed to get financial governance: %v", err)
	}
	if config.Thresholds != nil {
		fmt.Printf("Warning threshold: %d%%\n", config.Thresholds.WarningPercent)
		fmt.Printf("Critical threshold: %d%%\n", config.Thresholds.CriticalPercent)
		fmt.Printf("Exceeded action: %s\n", config.Thresholds.ExceededAction)
	}

	// Example 2: Get budget caps
	fmt.Println("\n=== Budget Caps ===")
	budgets, _, err := client.GetBudgetCaps(ctx, workspaceID)
	if err != nil {
		log.Fatalf("Failed to get budget caps: %v", err)
	}
	if budgets.Workspace != nil {
		fmt.Printf("Workspace budget: %.2f %s (used: %.2f)\n",
			budgets.Workspace.Limit,
			budgets.Workspace.Currency,
			budgets.Workspace.Used)
	}

	// Example 3: Update user spending limit
	fmt.Println("\n=== Update User Spending Limit ===")
	userID := int64(123) // Replace with actual user ID
	dailyLimit := 100.0
	_, _, err = client.UpdateUserSpendingLimit(ctx, workspaceID, userID, &management.UpdateUserSpendingLimitRequest{
		DailyLimit: &dailyLimit,
	})
	if err != nil {
		log.Printf("Failed to update spending limit: %v", err)
	} else {
		fmt.Printf("Updated daily limit for user %d: $%.2f\n", userID, dailyLimit)
	}

	// Example 4: Get spend summary
	fmt.Println("\n=== Spend Summary ===")
	summary, _, err := client.GetSpendSummary(ctx, workspaceID)
	if err != nil {
		log.Fatalf("Failed to get spend summary: %v", err)
	}
	if summary.Workspace != nil {
		fmt.Printf("Total workspace spend: %.2f %s\n",
			summary.Workspace.Spent,
			summary.Workspace.Currency)
	}
}
