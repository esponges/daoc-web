# Design: rules-sourcing-pass

## Context

- `internal/server/rules/registry.go` lists every rule with a status, and `cmd/parity` counts them: 26 documented, 66 reference, 43 provisional.
- The adopted sources are listed in CLAUDE.md and the README.

## Decisions

**D1. Order by how much players feel it.**
1. Melee and spell damage, resists, archery.
2. Experience and hit points.
3. Regeneration, endurance, resting and falling.
4. Class tables and item condition.

**D2. A registry flag.** Add `Searched string` (a date and a short note) to registry entries, and have the census count it.

**D3. A design log.** This file's "Log" section gets one line per rule touched, giving the old and new status and value. This is the record players can read when a number changes.

**D4. Never copy.** Reference servers can confirm what a number is in their world, but only a period source makes it documented.

## Log

(filled during the pass)
