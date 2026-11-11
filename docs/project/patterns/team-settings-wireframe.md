# Team settings (wireframe reference)

Design source: `dockyrd-user-settings.html` export.

HTML reference (team page + invite/role/remove modals): [`team-settings-wireframe.html`](./team-settings-wireframe.html)

Implemented route: `/settings/server/team` (`TeamSettingsPage`).

Flows:

- Invite by email / invite link
- Invites sent confirmation
- Change role & access
- Remove member

Project-scoped access in the wireframe is **not** wired to the API yet (shows “All projects” until v0.2 `project_members`).
