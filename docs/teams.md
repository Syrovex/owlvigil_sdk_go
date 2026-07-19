# Teams & Members Management Guide

> This guide provides extended workflows. For the complete source-domain map,
> role management, permission overrides, and workspace activity, see
> [Access control](access-control.md).

Complete guide to managing workspace teams and members with the OwlVigil Go SDK.

---

## Overview

The Teams & Members API allows you to:
- 👥 Create and manage teams
- 👤 Invite and manage members
- 📧 Send and manage invitations
- 🔐 Assign roles and permissions
- 📊 Track team membership

---

## Authentication

All team operations require a service-account API key with the appropriate scopes:

```go
client := management.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")))
```

Required scopes:
- `workspace:team:read` - List and view teams
- `workspace:team:write` - Create, update, delete teams
- `workspace:member:read` - List and view members
- `workspace:member:write` - Invite, update, remove members

---

## Teams

### List Teams

```go
teams, meta, err := client.ListTeams(ctx, workspaceID, management.ListOptions{
    Limit: 10,
})
if err != nil {
    log.Fatalf("Failed to list teams: %v", err)
}

for _, team := range teams.Items {
    fmt.Printf("Team: %s (%d members)\n", team.Name, team.MemberCount)
}
```

### Create Team

```go
team, _, err := client.CreateTeam(ctx, workspaceID, &management.CreateTeamRequest{
    Name:        "Engineering",
    Description: "Engineering team",
})
if err != nil {
    log.Fatalf("Failed to create team: %v", err)
}

fmt.Printf("Created team: %s (ID: %d)\n", team.Name, team.ID)
```

### Get Team Details

```go
team, _, err := client.GetTeam(ctx, workspaceID, teamID)
if err != nil {
    log.Fatalf("Failed to get team: %v", err)
}

fmt.Printf("Team: %s\n", team.Name)
fmt.Printf("Status: %s\n", team.Status)
fmt.Printf("Members: %d\n", team.MemberCount)
```

### Update Team

```go
name := "Engineering Team"
description := "Core engineering team"

updated, _, err := client.UpdateTeam(ctx, workspaceID, teamID, &management.UpdateTeamRequest{
    Name:        &name,
    Description: &description,
})
if err != nil {
    log.Fatalf("Failed to update team: %v", err)
}

fmt.Printf("Updated team: %s\n", updated.Name)
```

### Delete Team

```go
_, err := client.DeleteTeam(ctx, workspaceID, teamID)
if err != nil {
    log.Fatalf("Failed to delete team: %v", err)
}

fmt.Println("Team deleted successfully")
```

---

## Members

### List Members

```go
members, _, err := client.ListMembers(ctx, workspaceID, management.ListOptions{
    Limit: 20,
})
if err != nil {
    log.Fatalf("Failed to list members: %v", err)
}

for _, member := range members.Items {
    fmt.Printf("Member: %s (%s) - %s\n", member.Name, member.Email, member.Status)
}
```

### Get Available Roles

```go
roles, _, err := client.ListRoleOptions(ctx, workspaceID)
if err != nil {
    log.Fatalf("Failed to list roles: %v", err)
}

for _, role := range roles.Items {
    fmt.Printf("Role: %s (%s)\n", role.Name, role.Description)
}
```

### Invite Member

```go
member, _, err := client.CreateMember(ctx, workspaceID, &management.CreateMemberRequest{
    Email:   "newuser@example.com",
    RoleIDs: []int64{1, 2},  // Admin and Developer roles
    TeamIDs: []int64{5},      // Engineering team
})
if err != nil {
    log.Fatalf("Failed to invite member: %v", err)
}

fmt.Printf("Invited: %s (Status: %s)\n", member.Email, member.Status)
```

### Get Member Details

```go
member, _, err := client.GetMember(ctx, workspaceID, userID)
if err != nil {
    log.Fatalf("Failed to get member: %v", err)
}

fmt.Printf("Member: %s\n", member.Name)
fmt.Printf("Email: %s\n", member.Email)
fmt.Printf("Roles: %v\n", member.RoleIDs)
fmt.Printf("Teams: %v\n", member.TeamIDs)
```

### Update Member

```go
updated, _, err := client.UpdateMember(ctx, workspaceID, userID, &management.UpdateMemberRequest{
    RoleIDs: []int64{1, 3},  // Update roles
    TeamIDs: []int64{5, 6},  // Update teams
})
if err != nil {
    log.Fatalf("Failed to update member: %v", err)
}

fmt.Printf("Updated member: %s\n", updated.Email)
```

### Remove Member

```go
_, err := client.DeleteMember(ctx, workspaceID, userID)
if err != nil {
    log.Fatalf("Failed to remove member: %v", err)
}

fmt.Println("Member removed successfully")
```

---

## Invitations

### List Invitations

```go
invitations, _, err := client.ListInvitations(ctx, workspaceID, management.ListOptions{
    Limit: 10,
})
if err != nil {
    log.Fatalf("Failed to list invitations: %v", err)
}

for _, inv := range invitations.Items {
    fmt.Printf("Invitation: %s - %s (expires: %s)\n",
        inv.Email, inv.Status, inv.ExpiresAt)
}
```

### Create Invitation

```go
invitation, _, err := client.CreateInvitation(ctx, workspaceID, &management.CreateInvitationRequest{
    Email:     "user@example.com",
    RoleIDs:   []int64{2},
    TeamIDs:   []int64{5},
    ExpiresAt: "2026-07-13T00:00:00Z",  // 7 days from now
})
if err != nil {
    log.Fatalf("Failed to create invitation: %v", err)
}

fmt.Printf("Invitation sent to: %s\n", invitation.Email)
fmt.Printf("Invite link: %s\n", invitation.InviteLink)
```

### Resend Invitation

```go
_, err := client.ResendInvitation(ctx, workspaceID, invitationID)
if err != nil {
    log.Fatalf("Failed to resend invitation: %v", err)
}

fmt.Println("Invitation resent")
```

### Revoke Invitation

```go
_, err := client.RevokeInvitation(ctx, workspaceID, invitationID)
if err != nil {
    log.Fatalf("Failed to revoke invitation: %v", err)
}

fmt.Println("Invitation revoked")
```

---

## Common Workflows

### 1. Onboard New Team Member

```go
func onboardMember(ctx context.Context, client *management.Client, workspaceID int64, email string) error {
    // Step 1: Create invitation
    invitation, _, err := client.CreateInvitation(ctx, workspaceID, &management.CreateInvitationRequest{
        Email:   email,
        RoleIDs: []int64{2}, // Developer role
        TeamIDs: []int64{5}, // Engineering team
    })
    if err != nil {
        return fmt.Errorf("failed to create invitation: %w", err)
    }

    // Step 2: Send invite link via email (your app's logic)
    fmt.Printf("Send invitation to %s: %s\n", email, invitation.InviteLink)

    return nil
}
```

### 2. Reorganize Team Structure

```go
func reorganizeTeams(ctx context.Context, client *management.Client, workspaceID int64) error {
    // Step 1: Create new teams
    frontend, _, err := client.CreateTeam(ctx, workspaceID, &management.CreateTeamRequest{
        Name:        "Frontend",
        Description: "Frontend engineers",
    })
    if err != nil {
        return err
    }

    backend, _, err := client.CreateTeam(ctx, workspaceID, &management.CreateTeamRequest{
        Name:        "Backend",
        Description: "Backend engineers",
    })
    if err != nil {
        return err
    }

    // Step 2: Reassign members to new teams
    members, _, err := client.ListMembers(ctx, workspaceID, management.ListOptions{})
    if err != nil {
        return err
    }

    for _, member := range members.Items {
        // Example: assign based on some criteria
        newTeamIDs := []int64{frontend.ID} // or backend.ID
        _, _, err := client.UpdateMember(ctx, workspaceID, member.ID, &management.UpdateMemberRequest{
            TeamIDs: newTeamIDs,
        })
        if err != nil {
            log.Printf("Failed to update member %d: %v", member.ID, err)
        }
    }

    return nil
}
```

### 3. Bulk Invite Users

```go
func bulkInvite(ctx context.Context, client *management.Client, workspaceID int64, emails []string, roleIDs []int64) error {
    for _, email := range emails {
        invitation, _, err := client.CreateInvitation(ctx, workspaceID, &management.CreateInvitationRequest{
            Email:   email,
            RoleIDs: roleIDs,
        })
        if err != nil {
            log.Printf("Failed to invite %s: %v", email, err)
            continue
        }

        fmt.Printf("Invited: %s (Link: %s)\n", email, invitation.InviteLink)

        // Rate limiting
        time.Sleep(100 * time.Millisecond)
    }

    return nil
}
```

### 4. Clean Up Expired Invitations

```go
func cleanupInvitations(ctx context.Context, client *management.Client, workspaceID int64) error {
    invitations, _, err := client.ListInvitations(ctx, workspaceID, management.ListOptions{})
    if err != nil {
        return err
    }

    now := time.Now()
    for _, inv := range invitations.Items {
        if inv.Status == "pending" && inv.ExpiresAt != "" {
            expiresAt, err := time.Parse(time.RFC3339, inv.ExpiresAt)
            if err != nil {
                continue
            }

            if now.After(expiresAt) {
                _, err := client.RevokeInvitation(ctx, workspaceID, inv.ID)
                if err != nil {
                    log.Printf("Failed to revoke invitation %d: %v", inv.ID, err)
                } else {
                    fmt.Printf("Revoked expired invitation for %s\n", inv.Email)
                }
            }
        }
    }

    return nil
}
```

---

## Error Handling

### Check Member Existence

```go
member, _, err := client.GetMember(ctx, workspaceID, userID)
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        fmt.Println("Member does not exist")
        return
    }
    log.Fatalf("Unexpected error: %v", err)
}
```

### Handle Duplicate Invitations

```go
_, _, err := client.CreateInvitation(ctx, workspaceID, &management.CreateInvitationRequest{
    Email:   "user@example.com",
    RoleIDs: []int64{1},
})
if err != nil {
    if strings.Contains(err.Error(), "already invited") ||
       strings.Contains(err.Error(), "already member") {
        fmt.Println("User is already invited or is a member")
        return nil
    }
    return err
}
```

---

## Best Practices

### 1. Use Pagination for Large Teams

```go
cursor := ""
for {
    members, meta, err := client.ListMembers(ctx, workspaceID, management.ListOptions{
        Limit:  50,
        Cursor: cursor,
    })
    if err != nil {
        return err
    }

    // Process members
    for _, member := range members.Items {
        fmt.Printf("Member: %s\n", member.Email)
    }

    // Check if more pages
    if !meta.HasMore {
        break
    }
    cursor = meta.NextCursor
}
```

### 2. Validate Email Before Inviting

```go
import "net/mail"

func inviteWithValidation(ctx context.Context, client *management.Client, workspaceID int64, email string) error {
    // Validate email format
    _, err := mail.ParseAddress(email)
    if err != nil {
        return fmt.Errorf("invalid email format: %w", err)
    }

    // Check if already a member
    members, _, err := client.ListMembers(ctx, workspaceID, management.ListOptions{})
    if err != nil {
        return err
    }

    for _, member := range members.Items {
        if member.Email == email {
            return fmt.Errorf("user is already a member")
        }
    }

    // Create invitation
    _, _, err = client.CreateInvitation(ctx, workspaceID, &management.CreateInvitationRequest{
        Email:   email,
        RoleIDs: []int64{2},
    })

    return err
}
```

### 3. Set Expiration for Invitations

```go
// Invitation expires in 7 days
expiresAt := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)

invitation, _, err := client.CreateInvitation(ctx, workspaceID, &management.CreateInvitationRequest{
    Email:     "user@example.com",
    RoleIDs:   []int64{2},
    ExpiresAt: expiresAt,
})
```

### 4. Track Team Changes

```go
// Before updating team membership
originalTeams := member.TeamIDs

// Update
updated, _, err := client.UpdateMember(ctx, workspaceID, userID, &management.UpdateMemberRequest{
    TeamIDs: newTeamIDs,
})
if err != nil {
    return err
}

// Log changes
fmt.Printf("Member %s team membership changed:\n", member.Email)
fmt.Printf("  Before: %v\n", originalTeams)
fmt.Printf("  After: %v\n", updated.TeamIDs)
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
    workspaceID := int64(1) // Replace with your workspace ID

    client := management.NewClient(
        owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
    )

    // Create a new team
    team, _, err := client.CreateTeam(ctx, workspaceID, &management.CreateTeamRequest{
        Name:        "Engineering",
        Description: "Engineering team",
    })
    if err != nil {
        log.Fatalf("Failed to create team: %v", err)
    }
    fmt.Printf("Created team: %s (ID: %d)\n", team.Name, team.ID)

    // Invite a new member to the team
    member, _, err := client.CreateMember(ctx, workspaceID, &management.CreateMemberRequest{
        Email:   "engineer@example.com",
        RoleIDs: []int64{2}, // Developer role
        TeamIDs: []int64{team.ID},
    })
    if err != nil {
        log.Fatalf("Failed to invite member: %v", err)
    }
    fmt.Printf("Invited: %s to team %s\n", member.Email, team.Name)

    // List all teams
    teams, _, err := client.ListTeams(ctx, workspaceID, management.ListOptions{})
    if err != nil {
        log.Fatalf("Failed to list teams: %v", err)
    }

    fmt.Println("\nAll Teams:")
    for _, t := range teams.Items {
        fmt.Printf("  - %s (%d members)\n", t.Name, t.MemberCount)
    }

    // List all members
    members, _, err := client.ListMembers(ctx, workspaceID, management.ListOptions{})
    if err != nil {
        log.Fatalf("Failed to list members: %v", err)
    }

    fmt.Println("\nAll Members:")
    for _, m := range members.Items {
        fmt.Printf("  - %s (%s) - %s\n", m.Name, m.Email, m.Status)
    }
}
```

---

## See Also

- [Access control](access-control.md) - Roles and permissions management
- [Authentication Guide](./authentication.md)
- [Error Handling](./errors.md)
- [Examples](../examples/team-management/)
