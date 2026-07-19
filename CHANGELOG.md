# Changelog

## v0.2.0 (2026-07-06)

### Added - Management API (99 new endpoints)

#### Billing & Payment (28 endpoints)
- **Subscription Management (12 endpoints)**
  - `ListPlans()` - List subscription plans
  - `GetPlan()` - Get plan details
  - `GetSubscription()` - Get subscription status
  - `CreateSubscriptionCheckout()` - Create subscription checkout
  - `CreateSubscriptionInApp()` - In-app subscription payment
  - `ConfirmSubscriptionInApp()` - Confirm in-app subscription
  - `UpgradeSubscription()` - Upgrade subscription
  - `DowngradeSubscription()` - Downgrade subscription
  - `CancelSubscription()` - Cancel subscription
  - `ReactivateSubscription()` - Reactivate subscription
  - `GetSubscriptionCheckoutSession()` - Get checkout session status
  - `SyncLatestSubscriptionCheckout()` - Sync latest checkout

- **Top-up Management (6 endpoints)**
  - `ListTopupPlans()` - List top-up plans
  - `CreateTopupCheckout()` - Create top-up checkout
  - `CreateTopupInApp()` - In-app top-up
  - `ConfirmTopupInApp()` - Confirm in-app top-up
  - `ListTopups()` - List top-up transactions
  - `GetTopup()` - Get top-up details

- **Payment Methods (5 endpoints)**
  - `ListPaymentMethods()` - List saved payment methods
  - `CreatePaymentMethodSetupIntent()` - Create SetupIntent
  - `SavePaymentMethod()` - Save payment method
  - `SetDefaultPaymentMethod()` - Set default payment
  - `DeletePaymentMethod()` - Remove payment method

- **Billing & Orders (5 endpoints)**
  - `GetBillingOverview()` - Get billing overview
  - `GetBillingDetails()` - Get billing details
  - `UpdateBillingDetails()` - Update billing details
  - `GetInvoice()` - Get invoice details
  - `ListOrders()`, `GetOrder()`, `ConfirmStripeSession()` - Order management

#### Members & Teams (15 endpoints)
- **Members (6 endpoints)**
  - `ListMembers()` - List workspace members
  - `ListRoleOptions()` - List available roles
  - `CreateMember()` - Invite member
  - `GetMember()` - Get member details
  - `UpdateMember()` - Update member
  - `DeleteMember()` - Remove member

- **Teams (5 endpoints)**
  - `ListTeams()` - List teams
  - `CreateTeam()` - Create team
  - `GetTeam()` - Get team details
  - `UpdateTeam()` - Update team
  - `DeleteTeam()` - Delete team

- **Invitations (4 endpoints)**
  - `ListInvitations()` - List invitations
  - `CreateInvitation()` - Create invitation
  - `ResendInvitation()` - Resend invitation
  - `RevokeInvitation()` - Revoke invitation

#### RBAC (9 endpoints)
- `ListRoles()` - List roles
- `CreateRole()` - Create custom role
- `GetRole()` - Get role details
- `UpdateRole()` - Update role
- `DeleteRole()` - Delete role
- `ListPermissions()` - List permissions
- `GetMemberPermissions()` - Get member permissions
- `UpdateMemberPermissions()` - Update member permissions
- `ResetMemberPermissions()` - Reset member permissions

#### Financial Control (12 endpoints)
- `GetFinancialGovernance()` - Get financial config
- `UpdateFinancialGovernance()` - Update financial config
- `GetBudgetCaps()` - Get budget caps
- `UpdateBudgetCaps()` - Update budget caps
- `UpdateScopeBudgetCap()` - Update scope budget
- `GetSpendingLimits()` - Get spending limits
- `UpdateSpendingLimits()` - Update spending limits
- `UpdateUserSpendingLimit()` - Update user limit
- `GetFinancialThresholds()` - Get thresholds
- `UpdateFinancialThresholds()` - Update thresholds
- `PreviewFinancialChanges()` - Preview changes
- `GetSpendSummary()` - Get spend summary

#### User & Account (10 endpoints)
- `GetUserProfile()` - Get user profile
- `UpdateUserProfile()` - Update user profile
- `UpdatePassword()` - Update password
- `CreateSupportRequest()` - Submit support request
- `GetNotificationPreferences()` - Get notification settings
- `UpdateNotificationPreferences()` - Update notification settings
- `GetInviteLink()` - Get invite link
- `GetInvitationStats()` - Get invitation stats
- `ListUserInvitations()` - List invitations
- `SendInvitation()` - Send invitations

#### Models & Routes (4 endpoints)
- `ListModels()` - List models
- `GetModel()` - Get model details
- `ListRoutes()` - List routes
- `PreviewRoute()` - Preview routing

#### Policies (2 endpoints)
- `GetGatewayPolicies()` - Get gateway policies
- `PreviewPolicyEffect()` - Preview policy effect

#### Webhooks (11 endpoints)
- `GetWebhookEndpoint()` - Get endpoint details
- `UpdateWebhookEndpoint()` - Update endpoint
- `DeleteWebhookEndpoint()` - Delete endpoint
- `EnableWebhookEndpoint()` - Enable endpoint
- `DisableWebhookEndpoint()` - Disable endpoint
- `RotateWebhookSecret()` - Rotate secret
- `ListWebhookEventTypes()` - List event types
- `GetWebhookEvent()` - Get event details
- `ListEndpointEvents()` - List endpoint events
- `RetryWebhookEvent()` - Retry event
- `RedeliverWebhookEvent()` - Redeliver event
- `BulkRedeliverWebhookEvents()` - Bulk redeliver

#### Logs & Usage (8 endpoints)
- `ListTraces()` - List traces
- `GetPayloadAccess()` - Check payload access
- `GetPayloadLog()` - Get payload log
- `ListUsage()` - List usage records
- `GetQuotaSummary()` - Get quota summary
- `GetQuotaUsage()` - Get quota usage breakdown

#### Workspace (2 endpoints)
- `UpdateWorkspace()` - Update workspace
- `ListWorkspaceActivity()` - List activity logs

### Changed
- Enhanced error handling across all new endpoints
- Improved type definitions for better IDE support

### Notes
- All endpoints compile successfully
- 16 endpoints (11%) are placeholder implementations on server side
- Full API coverage: 141/141 endpoints (100%)

## v0.1.0 (2026-07-05)

### Added

- Bootstrap the standalone `github.com/Syrovex/owlvigil_sdk_go` SDK module.
- Add Gateway client for `https://api.owlvigil.com` model calls with Gateway key authentication.
- Add Management client for `https://open.owlvigil.com/open/v1` workflows with OAuth2.0 access-token authentication.
- Add OAuth2.0 helpers, streaming support, pagination helpers, webhook signature verification, examples, docs, and tests.
