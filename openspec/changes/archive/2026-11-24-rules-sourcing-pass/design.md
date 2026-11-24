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

The pass of 2026-11-24 searched 44 of the 58 provisional rules with three parallel research passes and found almost nothing numeric for 1.65; the open web has little, and the Herald and fandom class pages would not load. One number changed:

- `movement.sprint`: speed factor 1.25 to 1.30, from a Grab Bag answer (sprint is a 30 percent velocity bonus). The endurance cost stays provisional. The shared move vectors were rewritten.

Every other searched rule keeps its value; what was searched and found is in `internal/server/rules/sourcing.go` and shows in the census. Findings that disagree with ours but were not adopted (one source, later era or second hand): power regeneration (a 2009 test says about 15 a tick at level 50, ours 23), charm having no level cap, the haste floor of 1.5 seconds, Critical Shot needing stealth, and a group-challenge rule for experience. 14 provisional rules were not searched: the project's own features (guilds, crafting, tasks, ground target and the like), which have no 1.65 number to find.
