# Bulletin Project Instructions

Bulletin is a community platform for organizing discussions and real-time chat within "Circles".

## Project Overview

- **Backend**: Go (Golang) API.
- **Frontend**: Vue.js with Pinia for state management.
- **Styling**: Modular CSS in `frontend/src/styles/` using CSS Custom Properties and standard CSS nesting.
- **Documentation**: Detailed project specifications and API references are available in the `docs/` folder.

## Core Principles

- **No Tailwind**: Tailwind has been removed. Use the modular CSS system.
- **Semantic HTML**: Style by element and structure whenever possible.
- **Minimal Classes**: Keep HTML clean; avoid excessive class attributes.
- **Responsive Design**: Handle responsiveness via media queries in CSS files, not framework utilities.

## CSS Architecture

Styles are organized into specific modules in `frontend/src/styles/`:

- `variables.css`: All CSS custom properties and palette definitions.
- `base.css`: Element defaults (body, headers, links, code).
- `layout.css`: Major structural layouts (app-layout, circle-layout).
- `forms.css`: Labels, inputs, buttons, and form groupings.
- `components.css`: Reusable UI components (icons, cards, chat messages, modals).
- `markdown.css`: Styles for rendered markdown content.
- `animations.css`: Keyframes and Vue transition classes.

### Guidelines

- **No Inline Styles**: `style="..."` attributes are prohibited unless required for dynamic values.
- **No Utility Classes**: Avoid utility-first styling.
- **No Scoped Styles in Views**: View-level styles belong in global CSS modules. Components may use scoped styles for specific animations or transitions.
- **Theming**: 6 color palettes are supported via body classes (`palette-ocean|ember|forest|rose|slate`, violet is default). Use CSS custom properties (`var(--accent)`, `var(--bg-raised)`) to ensure compatibility.

## Interaction & Code Quality

- **Terse Responses**: Provide direct answers without recaps or summaries.
- **No Obvious Comments**: Only comment code when the "why" logic is non-obvious.
- **Specificity**: Maintain low CSS specificity by using element selectors and `:where()` when appropriate.
