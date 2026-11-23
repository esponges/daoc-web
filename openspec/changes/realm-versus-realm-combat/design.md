# Design: realm-versus-realm-combat

## Context

- **`mayHarm`** comes from `player-trade-and-social` (duels). This change adds "different realm, both in an RvR area".
- **Hit location:** the style feature "fixed hit location" is counted but unused, because nothing fought players. Player armour per slot already exists for creature attacks, so player against player reuses it.
- **Realm abilities data:** `classxrealmability_atlas` holds the Atlas server's set, which is close to classic. Each ability is checked against 1.65 sources.

## Decisions

**D1. RvR areas are a zone property.** `rvr: true` is set for the frontier zones (classic frontiers inside regions 1, 100 and 200, which task 1.1 lists) and the battleground regions.

**D2. Damage credit.** Each player keeps a damage-by-attacker map that clears on death or full health. It is the same structure creature experience credit uses.

**D3. Realm abilities as rule entries.** Each ability is a registry rule with its per-level amounts and status. Passive abilities fold into the stat and rule functions through the existing bonus sums. Active abilities are effects with reuse timers, as in `finish-midgard-classes` D2.

**D4. Immunity timers** are kept per player per kind in the effects list. They are not shown as icons, as in the game.

## Risks / Trade-offs

- [The 1.65 numbers for realm points and anti-farming are thin] → reference status first, then the sourcing pass.
- [Few players to test with] → admin commands to spawn test characters of each realm and to set realm points.
