# Members, Access Control, and Gateway Keys

## Access Control

Use the Management client with `OWLVIGIL_API_KEY`. These APIs change the
selected workspace or its members, so obtain the workspace ID first and keep
write operations behind an administrator-only action in your application.

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
workspaces, _, err := client.ListWorkspaces(ctx, management.ListOptions{Limit: 20})
if err != nil {
	return err
}
if len(workspaces.Items) == 0 {
	return errors.New("no OwlVigil workspace is available")
}
workspaceID := workspaces.Items[0].ID

team, _, err := client.CreateTeam(ctx, workspaceID, &management.CreateTeamRequest{
	Name: "platform",
})
if err != nil {
	return err
}
_ = team

```

### Workspaces

Use `ListWorkspaces` to select a workspace, then `GetWorkspace` to refresh its
details. `UpdateWorkspace` changes workspace settings and
`ListWorkspaceActivity` retrieves a cursor-paginated activity feed. Treat a
workspace ID as an authorization boundary: do not accept it from an untrusted
caller without checking that caller's own access first.

### Teams, members, and invitations

Teams are scoped to a workspace. `ListTeams`, `CreateTeam`, `GetTeam`,
`UpdateTeam`, and `DeleteTeam` manage teams. `ListMembers`, `CreateMember`,
`GetMember`, `UpdateMember`, and `DeleteMember` manage workspace members.
`ListRoleOptions` returns roles suitable for member assignment.

Invitations are also workspace-scoped:

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
invite, _, err := client.CreateInvitation(ctx, workspaceID,
	&management.CreateInvitationRequest{Email: "engineer@example.com"},
)
if err != nil {
	return err
}
_, err = client.ResendInvitation(ctx, workspaceID, invite.ID)
if err != nil {
	return err
}

```

Use `ListInvitations` for pagination, `ResendInvitation` only after confirming
the destination address, and `RevokeInvitation` when access is no longer
required. Deleting a member, team, or invitation is a write operation and may
affect running workflows; present a confirmation in interactive applications.

The older [Teams and members guide](06-management-access-and-keys.md#teams-and-members) contains extended onboarding and
bulk-invite examples. This page is the canonical API-domain map.

### Roles and permission overrides

`ListRoles`, `CreateRole`, `GetRole`, `UpdateRole`, and `DeleteRole` manage
custom roles. `ListPermissions` lists assignable permission identifiers.
`GetMemberPermissions`, `UpdateMemberPermissions`, and
`ResetMemberPermissions` manage a member-specific override.

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
role, _, err := client.CreateRole(ctx, workspaceID, &management.CreateRoleRequest{
	Name:        "auditor",
	Description: "Can inspect usage and request logs.",
	Permissions: []string{"usage:read", "logs:read"},
})
if err != nil {
	return err
}
_ = role

```

Prefer roles over per-member overrides when a policy applies to more than one
person. Before calling `UpdateMemberPermissions`, read the current override so
your interface does not unintentionally replace a permission chosen elsewhere.

### Pagination and auditability

`ListWorkspaces`, `ListWorkspaceActivity`, `ListTeams`, `ListMembers`,
`ListInvitations`, and `ListRoles` accept `management.ListOptions`. Follow
`PageInfo.NextCursor` as described in [Pagination](05-management.md#pagination). For every
write, retain the response metadata request ID; it is the fastest way to trace
an access change with support.

## Teams and Members

Teams, members, invitations, roles, and permission overrides are workspace-scoped Management resources. This page is a task map; [Access control](06-management-access-and-keys.md#access-control) contains the operational rules and [Go package documentation](https://pkg.go.dev/github.com/Syrovex/owlvigil_sdk_go/management) is the field-level reference.

### Current resource model

<!-- evidence: management/teams.go, management/members.go, management/rbac.go -->
```text
Workspace
├── Teams
├── Members
│   └── Member permission overrides
├── Invitations
└── Roles
    └── Permissions
```

Use the current singular fields on member requests: `Role` and `TeamID`. The older `RoleIDs` and `TeamIDs` fields remain only for source compatibility and should not be used in new integrations.

### Recommended workflow

1. Call `ListPermissions` and `ListRoles` to inspect the available access model.
2. Create or select a role and team.
3. Add a member with `CreateMember` or invite a user with `CreateInvitation`.
4. Read the member with `GetMember` and verify the effective access with `GetMemberPermissions`.
5. Use `UpdateMemberPermissions` only for an intentional per-member exception.
6. Use `ResetMemberPermissions` to remove an exception and return to the role baseline.

Team methods are `ListTeams`, `CreateTeam`, `GetTeam`, `UpdateTeam`, and `DeleteTeam`. Member, invitation, role, and permission methods are listed in [Management](05-management.md#management-api).

### Safety rules

- Treat email addresses, invite links, and membership activity as personal data.
- Do not automatically retry create, resend, revoke, update, or delete operations; these routes do not document idempotency-key support.
- Confirm the workspace ID and target member before a destructive operation.
- Prefer a shared role over repeated per-member overrides.
- Read the resulting resource after a permission change and retain the response request ID.

### Verified example

The runnable [`examples/team-management`](../../examples/team-management/) program lists teams and members and creates an invitation only when its explicit write flag is enabled. It is compiled by `go test ./examples/...` and its HTTP methods are covered by `management/teams_test.go` and `management/refactored_openapi_contract_test.go`.

### Result-returning action variants

Use `ResendInvitationWithResult`, `RevokeInvitationWithResult`, `EnableGatewayKeyWithResult`, `DisableGatewayKeyWithResult`, `DeleteGatewayKeyWithResult`, `DeleteMemberWithResult`, `DeleteTeamWithResult`, or `DeleteRoleWithResult` only when the server confirmation object is needed. Otherwise prefer the corresponding metadata-only method.
