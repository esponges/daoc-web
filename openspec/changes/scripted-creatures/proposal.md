# Proposal: scripted-creatures

## Why

The parity census counts 110 creature kinds in Midgard's data and called 23 of them implemented. Looking at what the server actually does showed the picture was wrong in both directions:

- Night and day creatures (NightSpawn, DaySpawn) already work, but the census called them unimplemented.
- The other kinds (Myling, ShiveringPresence, the named bosses and their adds, Enchanter, Recharger and others) spawn as plain creatures. The community data gives only a class type and a brain name for them, and public searches found nothing on what they do. A behaviour engine would have no rows.
- The real hole in the world is not behaviour but models: about 7% of Midgard's placements have no converted model, and none of the 35 models converts with the current character converter.

## What Changes

- **Census:** NightSpawn and DaySpawn count as implemented. Every other unimplemented kind carries its reason: owned by `frontiers-and-keeps` (keep guards), owned by `crafting` (crafting masters), an unsettled item service (Enchanter, Recharger), not classic (test dummies, other servers' teleporters, effect carriers), or behaviour unsourced.
- **Not-classic list:** gains the test dummies, the effect carrier, and the teleporters and trainers of other servers; the server already answers their interaction with a line.
- **Known limitation:** the README states how many Midgard placements lack a convertible model and why.
- **Deferred:** the behaviour layer (adds, flee, call for help, timed abilities). It waits for the first creature with a public source, and is written with that row.

## Capabilities

### New Capabilities
- `creature-behaviours`: time-of-day presence, and the rule that behaviour beyond a creature's template needs a source.

### Modified Capabilities
- `server-npcs`: dropped placements are accounted for in the README.

## Impact

- `cmd/parity` (reasons and the not-classic counting), `internal/importer/notclassic.go`, the README table and known limitations, `docs/parity`.
- No server behaviour changes.
