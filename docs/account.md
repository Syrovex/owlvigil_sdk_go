# Account Settings and Support

These Management methods operate on the authenticated user rather than a
Gateway key. Use `OWLVIGIL_API_KEY` and do not expose profile or invitation
data to another user without authorization.

## Profile and preferences

`GetUserProfile` reads the current profile; `UpdateUserProfile` updates only
the selected fields of `UpdateUserProfileRequest`. The current contract uses
`Username`; `Name` remains a compatibility alias. Set `ClearAvatarURL` or
`ClearBalanceNotifyThreshold` to send the explicit JSON `null` required to
clear those nullable values.

```go
name := "Ada Lovelace"
profile, _, err := client.UpdateUserProfile(ctx,
    &management.UpdateUserProfileRequest{Name: &name},
    owlvigil.WithIdempotencyKey("profile-name-ada"),
)
_ = profile
```

`GetNotificationPreferences` and `UpdateNotificationPreferences` manage budget,
billing, report, and marketing notification choices. The update is a PUT
replacement: read the current values first and submit all four choices. Any
unset request field becomes `false`.

## Password and support

`UpdatePassword` changes credentials using `UpdatePasswordRequest`. Handle the
password only in a protected form submission; do not log the request or return
it in an API response. `CreateSupportRequest` sends a typed support request
with a subject, issue type, and description. Include the SDK response request ID and a
sanitized reproduction in support tickets when possible.

Use `UpdatePasswordWithResult`, `CreateSupportRequestWithResult`, or
`SendInvitationWithResult` when the Open API action message or invitation
delivery count is needed. The original methods remain available for callers
that only need response metadata and an error.

## Invitations

`GetInviteLink` and `GetInvitationStats` retrieve the authenticated user's
invitation information. `ListUserInvitations` returns the complete published
list and accepts no pagination query; `SendInvitation` creates one. Invite
links can grant access, so do not place
them in public issues, analytics events, or client-side logs.

For workspace-admin invitations, member assignment, and role management, see
[Access control](access-control.md).
