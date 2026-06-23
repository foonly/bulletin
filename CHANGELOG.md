# Changelog

## 0.4.0 (2026-06-23)

#### Features

- chat: play notification sound on new message (80cf808)

## v0.3.0 (2026-06-13)

#### Features

- socket: migrate to global websocket connection (23ba125)

#### Refactor

- api: simplify loop syntax and optimize string building (a9fd521)

### v0.2.1 (2026-06-11)

#### Bug Fixes

- layout: redesign app header navigation with icons (e707ef6)

## v0.2.0 (2026-06-11)

#### Features

- version: display application version in UI (31ff796)
- ui: auto-focus input fields when opening reply/edit boxes and chat view (835f132)
- circle: add recent chat preview to dashboard (6a1cac5)
- circle: add invite management view and refine thread list UI (2654df3)
- header: add logo icon to app title (471ba06)
- circle: add palette support to circles (bf48787)
- circle: implement invite system and tag management features (98cfa66)
- tags: implement tag management features (b6c58fa)
- circles: add support for joining circles via invite links (9c0b54c)
- auth: make frontend base URL configurable via environment variable (d463790)
- mailer: add support for authenticated and TLS SMTP (6593258)
- notifications: add browser notification support for chat and activity (cad9750)
- posts: implement recursive unread count tracking for tags (2df6f9f)
- search: implement post search functionality (9662dc0)
- circle: implement full circle dashboard and navigation (cbb04e2)
- auth: implement user email verification, password reset, and TOTP MFA (6459a02)
- post: implement soft deletion for posts (300651b)
- chat: implement unread message tracking (82c8353)
- api: implement post updates and membership middleware (5a5e2ff)
- chat: implement background message cleanup and thread-based UI (d4a30d5)
- chat: improve WebSocket message handling and lifecycle (19fd442)

#### Bug Fixes

- css: Use CSS standard nesting. (3ae2a6a)
- circle: handle undefined chat messages in unread count calculation (ceed35a)

#### Refactor

- layout: standardize content width via content-container class (6439a23)
- style: nest CSS rules for modular consistency (ac80b6d)
- ui: remove inline styles and standardize layout classes (bbc11f8)
- styles: remove tailwind and implement modular css architecture (6190451)

#### Documentation

- readme: add reference to docs folder in AGENTS.md (1e728e2)
- agents: update project instructions and remove legacy memory files (c7f4a15)

#### Styles

- forms: increase left padding for search input fields (d8867e6)

#### Build System

- makefile: update build targets and add environment-based frontend URL (f1acb7f)
- makefile: add install target to copy binaries and frontend assets (49de034)
- frontend: migrate from npm to pnpm (bdce742)
- makefile: update build targets and improve help output (ce3f651)

#### Maintenance

- frontend: migrate from npm to pnpm (5398501)
- backend: remove temporary binary file (64b55b1)

### Misc
- Update .gitignore (7716464)
- Inital version (a178191)

