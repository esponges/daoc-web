# Design: battlegrounds

## Context

- **`battleground.sql`** gives each battleground's region, its level range and its realm-rank cap.
- **Keeps:** `keep.sql`, `keepcomponent.sql` and `keepposition.sql` describe keeps, their parts and the guard positions.
- **Keep models** are figures in the install. Task 1.1 finds how a keep is assembled from components (wall, tower, gate) and writes it up in `docs/formats.md`.

## Decisions

**D1. A keep is built from components on the server and the client.**
- The server knows the component positions and door entities.
- The client places the component figures.
- Collision uses the component hulls, as buildings do.

**D2. Doors are entities.** A door has health, a realm and a state. It is targetable and takes melee and spell damage through the normal combat paths, with its armour from the rule book (a `keep-doors` entry).

**D3. Guards reuse NPC combat.** Keep guard kinds become NPC templates with the owner realm and an aggressive-to-enemy-realm brain. Their numbers come from the keep's level.

**D4. Start with Thidranki.** Thidranki is built end to end first, and the other three follow with the same code.

## Non-Goals

- Siege, claiming and upgrades (`frontiers-and-keeps`).
