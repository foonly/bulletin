# Invite System & Permissions

Bulletin uses a double-tracked invite system to maintain a clear chain of origin for all users and their circle memberships.

## Invite Chain Logic

### Global Invite (`users.invited_by_id`)

When a user registers for the first time, the `invited_by_id` from the invite code's creator is saved to their user record. This represents the person who brought them into the Bulletin platform.

### Circle-Specific Invite (`circle_members.invited_by_id`)

Every time a user joins a circle (including their first one), the `invited_by_id` is tracked on the membership record. This allows a user to be invited to different circles by different people.

## Privacy Rules

To protect user privacy while maintaining traceability:

- **Shared Context**: When viewing a member list, you will see the name of the "Invited by" user ONLY if you share at least one circle with that inviter.
- **Anonymous fallback**: If there is no shared circle between you and the inviter, the UI displays "Unknown".
- **System Invites**: The initial bootstrap user (created via the `welcome` code) will show "System" as their inviter.

## Invite Interface

Invites are managed via a dedicated section in the sidebar and settings.

- **Generation**: A reusable `InviteModal` allows authorized users to create codes with specific roles and limits.
- **Tracking**: Admins can view "Active" and "Inactive/Depleted" invites.
- **Cleanup**: Expired or fully used codes can be cleared individually or in bulk.

## Role Permissions

| Action                      | Admin | Mod |    Standard    | Guest |
| :-------------------------- | :---: | :-: | :------------: | :---: |
| Edit Circle Settings        |  ✅   | ❌  |       ❌       |  ❌   |
| Manage Members (Roles/Kick) |  ✅   | ❌  |       ❌       |  ❌   |
| Manage Tags (Pin/Unpin)     |  ✅   | ✅  |       ❌       |  ❌   |
| Create Invites              |  ✅   | ✅  | Circle Setting |  ❌   |
| Delete Invites              |  ✅   | ✅  |       ❌       |  ❌   |
| Edit Others' Posts          |  ✅   | ❌  |       ❌       |  ❌   |
| Post Threads/Chat           |  ✅   | ✅  |       ✅       |  ✅   |

_Note: Circle settings can lower the requirement for "Create Invites" to allow Standard users to invite others._
