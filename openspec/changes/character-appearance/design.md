# Design: character-appearance

## Context

- **Heads today:** `charconv` dresses heads with the `Face` skin from `monsters.csv`. The `Head` column is the hair colour (CLAUDE.md), and the hair style is not mapped.
- **The pregame skin:** `uiconv` has a `pregame` skin folder (`web/data/ui/pregame`), which may hold the creation layout.

## Decisions

**D1. Find the hair in the install first** (task 1.1), with `nifdump` and `mpakls`. Hair is expected as separate meshes per race and gender, attached to the head bone, with textures tinted per colour. The findings are written into `docs/formats.md` before any code.

**D2. Appearance is one packed record:** `{face, hair, hairColour, eyes}` on the character and on the entity snapshot. NPCs fill it from their template's columns.

**D3. Stat points are stored as the distribution,** not the final stats, so the stat rules can recompute and validate. Old characters get an empty distribution.

**D4. The creation UI.** If the pregame layout converts, it is built from that. Otherwise the plain page gets pickers with a live preview (`charshot`'s renderer path).
