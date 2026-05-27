---
name: user-preferences
description: Niklas's coding and design preferences for this project
metadata:
  type: user
---

- **No style attributes** — inline `style="..."` attributes are strictly prohibited unless absolutely necessary for dynamic values. Prefer CSS classes.
- **Avoid utility classes** — style by element and structure. If a one-off style is needed, a utility class is preferred over an inline style attribute, but both should be avoided.
- **Semantic HTML-first styling** — minimal class attributes in HTML.
- **No scoped styles in views** — view-level styles go in the global CSS modules. Components may use scoped styles.
- **No Tailwind** — use CSS custom properties and modular CSS files.
- **Responsive design in CSS** — media queries in CSS, not framework utilities.
- **Terse responses** — no end-of-turn summaries, no trailing recap.
- **No comments in code** unless the WHY is non-obvious.
