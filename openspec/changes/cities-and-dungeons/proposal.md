# Proposal

## Why

295 of the game's 486 zones are cities and dungeons, and they use layouts the pipeline does not read at all: there is no heightmap, only assembled pieces. Until they load, "the whole world" means outdoor terrain only. Both layouts turn out to be plain text, so the cost is mostly new placement readers on top of the NIF reader we already have.

## What Changes

- Read dungeon layouts: `dungeon.chunk` (the list of piece models, for example `Cave_T.NIF`), `dungeon.place` (piece index, position, axis-angle rotation and flags per placed piece), `dungeon.prop` (props within the dungeon), with pieces from `zones/Dnifs`.
- Read city layouts: `city.csv` (the city's block models, for example Jordheim's `jord01.nif`–`jord31.nif` in `zones/zone120/nifs`), using `adjacency.csv` only if needed for visibility.
- Convert both into the same baked, instanced format the viewer already draws for scenery, plus each zone's lights and sector settings.
- Jordheim (zone 120) and one dungeon first, then every city and dungeon the census marks convertible; realm clones (for example "Cetus' Pit - Mid clone") reuse their source's conversion.
- The viewer loads any converted zone by number, with a camera suited to interiors (no sky, no terrain, start at the zone's entrance) and the same third-person character.
- Census rows for city and dungeon zones report convert status.

## Capabilities

### New Capabilities
- `dungeon-layout`: reading a dungeon zone's piece list and placements into renderable geometry.
- `city-layout`: reading a city zone's block models into renderable geometry.
- `zone-loading`: the viewer loading any converted zone by number, choosing outdoor or interior presentation by zone type.

### Modified Capabilities

## Impact

- New converter (or a mode of `propconv`) for dungeon and city layouts; shared baking code with outdoor scenery.
- Viewer: zone selection, interior presentation path (no terrain, sky, grass or ground detail), entrance spawn points.
- Depends on `world-census` for the parse status of `zones/Dnifs` and city block models; NIF block types closed there are prerequisites for full coverage.
- Out of scope: zone-to-zone transitions and collision.
