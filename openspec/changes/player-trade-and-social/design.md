# Design: player-trade-and-social

## Context

- **Commands** dispatch in `world/world.go`: `who`, `send`, `g`, `invite`, `accept` and `decline` exist. The `accept`/`decline` pair already serves group invites and resurrection offers.
- **Bags:** `world/inventory.go` owns them, and a vault move already changes two containers in one step.
- **Combat:** `world/combat.go` and `combatant.go` resolve player against creature. A player already has a combatant view (creatures attack players), so player against player reuses the defender side that creatures already use.
- **Stock layout:** it has `Trade`, `Friends` and `MiniFriends` windows.

## Decisions

**D1. One pending-offer mechanism.** Trade and duel prompts join the same accept/decline offers as group invites. Each offer carries a kind and a 30-second expiry.

**D2. Trade state lives on both players.**
- A `trade` struct is shared by pointer: two offers, two accepts, and a version.
- Each change bumps the version and clears the accepts.
- The commit checks both bags and purses again under the world lock, then moves everything or nothing.

**D3. Duels use a "may harm" predicate.**
- `mayHarm(a, b)` is true for a creature against a player, and for a player against a player in a duel with them.
- Later, `realm-versus-realm-combat` adds "different realm in an RvR zone".
- All attack paths (melee, styles, spells, archery, pets, area spells) go through it.

**D4. The friends and ignore lists** are stored in `players.db` as two small tables keyed by character ID. Login and logout notify the friends who are online.

**D5. Follow** is a server-side input source: each tick the server feeds the follower synthetic keys toward the target, and the client predicts from the keys it is sent. *Alternative:* a client-only follow. Rejected because the server owns movement.

## Risks / Trade-offs

- [The PvP combat path shares code with creature combat] → table tests for player defenders in melee, spell and archery before duels go on.
