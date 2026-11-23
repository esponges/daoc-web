# Design: frontiers-and-keeps

## Context

- **The classic frontiers** lie inside the realm regions (1, 100, 200), not in the later New Frontiers region. Their zones already convert with `worldconv`.
- **`keep-core`** comes from `battlegrounds`: doors, guards, lord and ownership.
- **Data:** `keephookpoint` and `keephookpointitem` give hookpoints and what can be bought there. `keepdoorteleport` gives the teleport links.

## Decisions

**D1. Keep levels and costs** go in a `keeps` rule-book entry: upgrade cost and time, door and guard strength by level. Status as sources allow.

**D2. Siege engines are entities with an operator.**
- An engine is placed at a ground target and has health.
- Aiming is a ground target for catapults and trebuchets, or a door for a ram.
- Firing is a timed action, and damage goes through the door and wall health.

**D3. Darkness Falls access** is computed from keep ownership on each change, and broadcast.

**D4. Order.**
1. Midgard's frontier zones and keeps.
2. Then the other realms' frontiers.
3. Then siege.
4. Then claiming and upgrades.
5. Then Darkness Falls.

## Non-Goals

- Relics (`relics`).
- New Frontiers (1.70+).
