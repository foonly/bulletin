---
name: project-css-architecture
description: Frontend CSS architecture after Tailwind removal — modular CSS files, custom properties, 6 palettes, dark/light mode
metadata:
  type: project
---

Tailwind removed (May 2026). Replaced with modular CSS in `frontend/src/styles/`:

- `index.css` — barrel import
- `variables.css` — all CSS custom properties + 6 palette classes
- `reset.css` — minimal reset
- `base.css` — element defaults (body, h1-h3, a, code, pre)
- `layout.css` — app-layout, circle-layout, auth-page, home-dashboard, etc.
- `forms.css` — label, input, textarea, select, .btn variants, .field
- `components.css` — circle-icon, card, thread-item, thread-node, chat, modal, toast, badges, palette-picker
- `markdown.css` — .markdown-content styles
- `animations.css` — @keyframes + Vue transition classes

**Palettes:** body class `palette-ocean|ember|forest|rose|slate` (violet is default, no class). Each palette overrides `--accent`, `--accent-*`, and surface colors (`--bg`, `--bg-raised`, etc.). Dark mode overrides are co-located in `@media (prefers-color-scheme: dark)` blocks. Circle admins select palette in CircleSettings; stored in localStorage as `circle-palette-<circleId>`. `Home.vue:applyPalette()` reads localStorage and applies the class to `document.body` on circle switch.

**Why:** User wanted zero Tailwind, semantic HTML, responsive CSS, and theming system for circle-level palette selection.

**How to apply:** When adding new UI, use CSS custom properties (`var(--accent)`, `var(--bg-raised)`, etc.) and add structural classes to the appropriate CSS file. Views have no scoped styles; components may use scoped styles for animations/transitions.
