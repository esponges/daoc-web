# Design: relics

## Context

- **`relic.sql`** gives six relics: positions, realm, original realm and type (0 Power, 1 Strength).
- **Keeps** come from `frontiers-and-keeps`, which owns keep ownership and the war map.

## Decisions

**D1. A relic is an entity with a state:** on pad, carried by a player, or on the ground. The state is kept in `world.db`'s runtime table, so a restart keeps it.

**D2. Gate rules from 1.65 sources.** Which keeps guard which temple is written into the `relics` rule-book entry with its source, along with the return time and the bonus per relic. No number is guessed: where no source gives one it is provisional and stated.

**D3. The carrier restrictions** are an effect on the carrier: speed effects are suppressed, teleports refused, and a war-map flag is set.

## Non-Goals

- Minotaur relics, which came later.
