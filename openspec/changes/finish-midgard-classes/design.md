# Design: finish-midgard-classes

## Context

- **Spell kinds** are dispatched in `world/pcasting.go` and `world/buffs.go`, and their numbers come from the community database via `world/catalog_spells.go`.
- **Abilities** are in `world/abilities.go` (Guard, Protect, Intercept, Engage, Camouflage and so on), with rules in `rules/abilities.go`.
- **Falling:** `rules/falling.go` already takes a Safe Fall level from Stealth spec. Only `cmd/parity`'s `gapReasons` and its implemented set are wrong.
- **Stock commands:** the client's command table carries `GroundTarget` (the stock key binding), and the HUD has no ground-target marker.

## Goals / Non-Goals

**Goals:**
- Every Midgard class can use what it trains to 50.
- The ability census reads 28 of 30 or better, and every remaining gap has a reason.

**Non-Goals:**
- Realm abilities (the `realm-versus-realm-combat` change).
- Climb Walls on keep walls (`frontiers-and-keeps`).

## Decisions

**D1. Paying in health.** Spells get a cost kind, power or health. The Savage kinds carry health, read from the database's `Power` column. Following the community data, a negative power value means a share of health. Task 1.1 confirms this by reading the data and the reference servers, without copying code. The rule book records it.

**D2. Timed stances are effects.** Berserk, Charge and QuickCast are effects in the existing effects list (`world/effects.go`), shown on the effects row with icons, and they expire on the shared tick. Their reuse timers sit in the same table that Engage and Camouflage already use.

**D3. The ground target is server state.**
- The client sends `groundTarget {x,y,z}`.
- The server checks range (2000) and line of sight with the existing `sight.go`, then stores it on the player.
- It is sent back as `gt` in the player's own state.
- *Alternative:* keep the target on the client only. Rejected because the server must aim at it.

**D4. Danger Sense** runs on the stealth tick, against hidden players of another realm. Until realm-versus-realm play exists nothing triggers it, but a test with a fake hidden player covers it.

**D5. Unknowns stay recorded.** If no source names "Call of a thousand storms" or result 3, they stay as gaps with a reason. They are not guessed.

## Risks / Trade-offs

- [Savage numbers from a reference server] → mark them reference or provisional, as the rule book requires.
- [The health cost could kill] → refuse when cost ≥ current health (spec scenario).
