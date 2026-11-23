# Proposal: battlegrounds

## Why

Battlegrounds were 1.65's level-banded RvR: small maps with one central keep, entered through each realm's portal keep, and closed to players over the level or realm-rank cap. The community data lists them in `battleground.sql`:

| battleground | levels |
| --- | --- |
| Thidranki | 20–24 |
| Murdaigean | 25–29 |
| Caledonia | 34–39 |
| Cathal Valley | 45–49 |

They are the smallest complete RvR loop, which makes them the right first step before the frontiers. They are also how a group of friends at level 20 first meets the other realms.

This change depends on `realm-versus-realm-combat` and on at least one other realm being playable (`albion-realm`).

## What Changes

- **Battleground zones are served:** terrain, the central keep, the portal keeps and the population of each battleground region. The convert and census steps are checked by eye.
- **Entry:**
  - each realm's portal NPC in its border keep sends a player of the right level and realm rank to their realm's portal keep in the battleground;
  - players over the cap are refused;
  - a player who levels past the cap while inside is removed.
- **The central keep:**
  - doors that can be destroyed;
  - a lord and guards from the battleground's data;
  - capture by killing the lord, which turns the keep to the capturing realm;
  - claiming by a guild is left to `frontiers-and-keeps`.

  This change brings in the keep machinery in its simplest form: doors, guards, a lord, ownership. `frontiers-and-keeps` extends it.
- **Realm points** are awarded as in RvR, and the battleground caps them at its realm-rank limit.
- **Census:** battlegrounds implemented out of 5 (Midgard's data counts 5).

## Capabilities

### New Capabilities
- `battlegrounds`: entry and level caps, portal keeps, the central keep's doors, guards, lord and capture.
- `keep-core`: keep doors (health, destruction, repair later), keep guards and lord, ownership by realm. Shared with `frontiers-and-keeps`.

### Modified Capabilities
- (none)

## Impact

- **Converters:** `worldconv` for the battleground regions, plus keep models (task 1.1 finds the keep component figures).
- **Import:** `keep`, `keepcomponent` and `keepposition`, and the battleground rows; re-import.
- **Server:** new `world/keeps.go` and `world/battlegrounds.go`; the keep guard kinds (`Keeps.GuardFighter`, `GuardArcher`, `GuardLord`, and the others in the census).
- **Client:** keep doors as attackable objects; the stock Keep window for status.
