# Access Control

Use the Management client with `OWLVIGIL_API_KEY`. These APIs change the
selected workspace or its members, so obtain the workspace ID first and keep
write operations behind an administrator-only action in your application.

```go
workspaces, _, err := client.ListWorkspaces(ctx, management.ListOptions{Limit: 20})
if err != nil || len(workspaces.Items) == 0 {
    return err
}
workspaceID := workspaces.Items[0].ID

team, _, err := client.CreateTeam(ctx, workspaceID, &management.CreateTeamRequest{
    Name: "platform",
}, owlvigil.WithIdempotencyKey("team-platform-20260713"))
```

## Workspaces

Use `ListWorkspaces` to select a workspace, then `GetWorkspace` to refresh its
details. `UpdateWorkspace` changes workspace settings and
`ListWorkspaceActivity` retrieves a cursor-paginated activity feed. Treat a
workspace ID as an authorization boundary: do not accept it from an untrusted
caller without checking that caller's own access first.

## Teams, members, and invitations

Teams are scoped to a workspace. `ListTeams`, `CreateTeam`, `GetTeam`,
`UpdateTeam`, and `DeleteTeam` manage teams. `ListMembers`, `CreateMember`,
`GetMember`, `UpdateMember`, and `DeleteMember` manage workspace members.
`ListRoleOptions` returns roles suitable for member assignment.

Invitations are also workspace-scoped:

```go
invite, _, err := client.CreateInvitation(ctx, workspaceID,
    &management.CreateInvitationRequest{Email: "engineer@example.com"},
    owlvigil.WithIdempotencyKey("invite-engineer-example"),
)
if err != nil {
    return err
}
_, err = client.ResendInvitation(ctx, workspaceID, invite.ID)
```

Use `ListInvitations` for pagination, `ResendInvitation` only after confirming
the destination address, and `RevokeInvitation` when access is no longer
required. Deleting a member, team, or invitation is a write operation and may
affect running workflows; present a confirmation in interactive applications.

The older [Teams and members guide](teams.md) contains extended onboarding and
bulk-invite examples. This page is the canonical API-domain map.

## Roles and permission overrides

`ListRoles`, `CreateRole`, `GetRole`, `UpdateRole`, and `DeleteRole` manage
custom roles. `ListPermissions` lists assignable permission identifiers.
`GetMemberPermissions`, `UpdateMemberPermissions`, and
`ResetMemberPermissions` manage a member-specific override.

```go
role, _, err := client.CreateRole(ctx, workspaceID, &management.CreateRoleRequest{
    Name:        "auditor",
    Description: "Can inspect usage and request logs.",
    Permissions: []string{"usage:read", "logs:read"},
})
_ = role
```

Prefer roles over per-member overrides when a policy applies to more than one
person. Before calling `UpdateMemberPermissions`, read the current override so
your interface does not unintentionally replace a permission chosen elsewhere.

## Pagination and auditability

`ListWorkspaces`, `ListWorkspaceActivity`, `ListTeams`, `ListMembers`,
`ListInvitations`, and `ListRoles` accept `management.ListOptions`. Follow
`PageInfo.NextCursor` as described in [Pagination](pagination.md). For every
write, retain the response metadata request ID; it is the fastest way to trace
an access change with support.
