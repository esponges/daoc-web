# Design: albion-realm

## Context

- **Midgard is hard-wired in places:**
  - regions 100, 101 and 151 in `dbimport` and `cmd/parity`;
  - the start table in `rules/start.go`;
  - class catalogues in `world/catalog*.go`;
  - creation in `web/login.js`.
- **Albion is in the install and the converters:** `worldconv` converts all 118 zones, including Albion's, and the census measures them.

## Decisions

**D1. Realm is a parameter, not a fork.** Every Midgard-only constant becomes a per-realm table: regions, start towns, trainers' class lists, census sections, and name-plate colours. Midgard's tests keep passing unchanged.

**D2. Classes by shared kind.** Most Albion lines map to spell and ability kinds the server already has, since the census counts by kind. The census lists what is new:
- the Necromancer pet and shade form;
- the Cabalist's simulacrum pets;
- the Theurgist's pets;
- the Friar's staff styles;
- the Minstrel's songs and charm.

Each new kind gets a rule-book entry.

**D3. Shade form is an effect.**
- It is an effect on the Necromancer while the pet lives.
- `mayHarm` redirects to the pet.
- Casting uses the pet as origin for range and sight.

**D4. Look at every zone.** As with Midgard, each Albion zone gets a browser check of its terrain, placements, interiors and exits before it is called done. The `fixtures.csv` issue (no axis-angle, Z 0) noted in the README is checked for Albion's zones.

## Non-Goals

- Realm-versus-realm combat itself (`realm-versus-realm-combat`).
