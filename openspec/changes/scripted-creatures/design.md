# Design: scripted-creatures

## Context

- The importer reads `ClassType` from the community spawn data, and the server spawns only the kinds the census marks implemented (`GameNPC`, merchants, guards, trainers and so on).
- Night-only presence already exists for creatures flagged by the database's time columns.
- The community server implements these kinds as C# scripts, which are GPL. They may be read to learn what a creature does. Nothing is copied.

## Goals / Non-Goals

**Goals:**
- Every Midgard placement is in the world or deliberately left out.
- Named camps behave like their 1.65 selves where a source says how.

**Non-Goals:**
- Keep guards (`frontiers-and-keeps`), crafting masters (`crafting`), guild NPCs (`player-guilds`).
- Raid encounters outside Midgard.

## Decisions

**D1. Behaviours as data.**
- A table maps kind name to behaviours:
  `{presence: night|day, adds: [{atPct, template, count}], flee: {belowPct}, callHelp: {range}, timed: [{every, spellId|styleId}], cycle: {...}}`.
- It lives in `world/behaviours.go` beside its tests.
- A behaviour runs on the NPC brain's existing tick.

**D2. Sources per kind.** Each row of the table cites a source: a Herald or Allakhazam bestiary entry, or the reference server read for understanding, with status reference. A row without a source spawns plain (spec "Sourced or marked").

**D3. Adds use templates from the database.** The add kinds (`StripeAdd` and others) have their own templates in the spawn data. A boss's wave spawns those templates.

**D4. Services.** The Recharger uses the existing services path (`world/services.go`), like the healer and the smith. The Enchanter, NameRegistrar and RoyalTreasuryClerk are researched first (task 1.3) and implemented or marked not classic.

## Risks / Trade-offs

- [Named boss behaviour is thinly documented] → spawn plain and mark it. A plain spawn still gives the camp and the loot.
