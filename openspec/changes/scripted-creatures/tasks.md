# Tasks

## 1. Research

- [x] 1.1 For each unimplemented Midgard kind, read its rows in the community data. (The data holds a class type and a brain name only; see `design.md` Context. Every kind's status is in the census.)
- [x] 1.2 Find a public source for each named creature's behaviour. (Searched 2026-11-23: none found for Myling, Queen Major, Stripe, Zrit-Zrit or Snarls; see `design.md`.)
- [x] 1.3 Decide whether Enchanter, NameRegistrar, RoyalTreasuryClerk and Recharger are classic 1.65. (Unsettled: no source; they spawn as plain creatures with no service, and the census says so.)

## 2. Behaviour layer

- [x] 2.1 Deferred by decision D1: no source gives any named creature's behaviour, so there is nothing for an engine to run. Revisit with the first sourced creature.
- [x] 2.2 Night and day presence: already implemented (`importer.SpawnTiming`, `world/timed.go`). The census now counts NightSpawn and DaySpawn as implemented.

## 3. Kinds

- [x] 3.1 to 3.4 Replaced by D2: every kind is classified in the census (presence, not classic, owned by a later change, unsettled service, or behaviour unsourced), and the test dummies, other servers' teleporters and effect carriers join the not-classic list.

## 4. Import and census

- [x] 4.1 Measure what the import drops: 1,693 Midgard placements across 35 models lack a converted model, and none of the 35 converts (design Context). Recorded in the README's known limitations.
- [x] 4.2 `cmd/parity`: new reasons and statuses; docs and the README table regenerated.

## 5. Check

- [x] 5.1 `go test -short ./...` passes. Nothing to play: the change adds no behaviour a player can see, so there is no browser check.
