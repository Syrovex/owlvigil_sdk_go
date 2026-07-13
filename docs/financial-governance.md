# Financial Governance

Financial controls are workspace-scoped Management writes. Require
`OWLVIGIL_API_KEY`, surface a confirmation to a human operator, and retain the
returned request ID. Use [Billing](billing.md) for payment and subscription
workflows.

## Inspect before changing

Read the current state with `GetFinancialGovernance`, `GetBudgetCaps`,
`GetSpendingLimits`, `GetFinancialThresholds`, and `GetSpendSummary`. The
spending-limits call is paginated. `GetQuotaSummary`, `GetQuotaUsage`,
`GetUsageSummary`, `ListUsage`, and `GetBalance` provide complementary quota,
usage, and balance evidence.

```go
governance, meta, err := client.GetFinancialGovernance(ctx, workspaceID)
if err != nil {
    return err
}
log.Printf("loaded financial governance request_id=%s", meta.RequestID)
_ = governance
```

## Set controls

`UpdateFinancialGovernance` can update budget caps, spending limits, and
thresholds together. Use the more focused `UpdateBudgetCaps`,
`UpdateSpendingLimits`, or `UpdateFinancialThresholds` when changing one
control family. `UpdateScopeBudgetCap` changes an individual workspace, team,
member, or Gateway-key scope; `UpdateUserSpendingLimit` changes one user's
limit.

```go
warning, critical := 80, 95
thresholds, _, err := client.UpdateFinancialThresholds(ctx, workspaceID,
    &management.UpdateThresholdsRequest{
        WarningPercent:  &warning,
        CriticalPercent: &critical,
    },
    owlvigil.WithIdempotencyKey("thresholds-80-95"),
)
_ = thresholds
```

Only populate pointer fields you intend to change. Before a broad update, call
`PreviewFinancialChanges` with a `PreviewFinancialChangesRequest`; it lets you
review the prospective result before persisting it. Avoid issuing automatic
budget changes based solely on a transient usage sample.

## Monitoring and evidence

Use `GetSpendSummary` for a workspace-level summary and `ListUsage` for
cursor-paginated records. `ListRequestLogs`, `GetRequestLog`, `ListTraces`, and
`GetTrace` connect spend to individual requests. Payload access and payload-log
methods may expose sensitive request content: request them only for an approved
incident and follow your data-retention policy.

See [Pagination](pagination.md) for list traversal and [Errors](errors.md) for
handling quota, authorization, and rate-limit failures.
