# Proposal: hair-styles

## Why

`character-appearance` split hair from the head and let a player pick a face and a hair colour. Only the first head mesh of each race (`HeadA1`) is converted, so every character of a race has one hair style. The install holds ten head meshes per race and gender (`HeadA1` to `A5`, `HeadB1` to `B5`), each a different hair style on the same face layout.

## What Changes

- `charconv` converts all ten head meshes per race and gender, splitting hair from face as for `HeadA1`.
- The creation page offers a hair style; the server stores and validates it; others see it.
- Eyes stay painted into the face skin; there is no eye colour to offer.
- Check the face/hair split on the races where it failed (trolls, kobolds, Valkyn, Frostalf).

## Capabilities

### Modified Capabilities
- `player-appearance`: a chosen hair style.

## Impact

`cmd/charconv`, the appearance table, `players.db` (a `hair_style` column), protocol (additive), the creation page.
