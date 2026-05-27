# Agent Instructions

Welcome to the Bulletin project. Please follow these guidelines:

## Core Principles
- **No Tailwind**: Use modular CSS in `frontend/src/styles/`.
- **Semantic HTML**: Style by element and structure.
- **Minimal Classes**: Keep HTML clean.

## CSS Guidelines
- **No inline styles**: `style="..."` attributes are strictly prohibited unless absolutely necessary for dynamic values.
- **No utility classes**: Avoid utility-first styling. If a one-off style is needed, a utility class is preferred over an inline style attribute, but both should be avoided.
- **No scoped styles in views**: View-level styles belong in global CSS modules. Components may use scoped styles for specific animations/transitions.

## Interaction
- **Terse responses**: No summaries or recaps.
- **No obvious comments**: Only comment non-obvious "why" logic.

For more details, see:
- `memory/MEMORY.md` (Index)
- `memory/user_preferences.md`
- `memory/project_css_architecture.md`
