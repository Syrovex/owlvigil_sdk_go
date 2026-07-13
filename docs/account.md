# Account Settings and Support

These Management methods operate on the authenticated user rather than a
Gateway key. Use `OWLVIGIL_API_KEY` and do not expose profile or invitation
data to another user without authorization.

## Profile and preferences

`GetUserProfile` reads the current profile; `UpdateUserProfile` updates only
the non-nil fields of `UpdateUserProfileRequest`. The request supports name,
avatar URL, and default workspace ID.

```go
name := "Ada Lovelace"
profile, _, err := client.UpdateUserProfile(ctx,
    &management.UpdateUserProfileRequest{Name: &name},
    owlvigil.WithIdempotencyKey("profile-name-ada"),
)
_ = profile
```

`GetNotificationPreferences` and `UpdateNotificationPreferences` manage budget,
billing, report, and marketing notification choices. Show the current values
before saving a replacement so users can make an informed change.

## Password and support

`UpdatePassword` changes credentials using `UpdatePasswordRequest`. Handle the
password only in a protected form submission; do not log the request or return
it in an API response. `CreateSupportRequest` sends a typed support request
with a title and description. Include the SDK response request ID and a
sanitized reproduction in support tickets when possible.

## Invitations

`GetInviteLink` and `GetInvitationStats` retrieve the authenticated user's
invitation information. `ListUserInvitations` paginates existing invitations;
`SendInvitation` creates one. Invite links can grant access, so do not place
them in public issues, analytics events, or client-side logs.

For workspace-admin invitations, member assignment, and role management, see
[Access control](access-control.md).
