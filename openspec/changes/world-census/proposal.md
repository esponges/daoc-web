# Proposal

## Why

The project has shown that DAoC's world can be rendered in a browser, but only for one zone (Vale of Mularn, 1 of 486). To claim the whole game's world loads, we need a measured answer — "X of Y assets parse and convert" across the entire install — and to close the gaps that measurement exposes. A first sample already shows the gaps are concentrated: 1,126 of 1,343 outdoor scenery models parse (84%), and seven unmodelled NIF block types account for 213 of the 217 failures.

## What Changes

- A census tool that walks the whole install and reports, per asset class, how many items parse and convert and why each failure fails:
  - zones by type from `zones.dat` (132 outdoor, 4 city, 291 dungeon, 19 other)
  - outdoor scenery models (`zones/Nifs`), dungeon pieces (`zones/Dnifs`, 3,681 archives), city blocks (`zones/zoneNNN/nifs`)
  - figures (`figures/`), animations (`anims/*.kfa`), and texture formats
- The report groups failures by cause (for example "unmodelled block type NiGravity: 64 files") and is regenerated from the install, never hand-edited; it is committed as a summary without game data.
- Close the NIF block types the census finds, starting with the seven known: `NiTextureEffect`, `NiGravity`, `NiTextureTransformController`, `NiGeomMorpherController`, `NiPixelData`, `NiAlphaController`, `NiMaterialColorController` — at minimum read exactly so files parse; render where the effect matters (for example `NiGravity` on particles, embedded `NiPixelData` textures).
- Convert every outdoor (type 0) zone with the existing `zoneconv` and `propconv`, recording which convert cleanly.
- A test that fails if the parse rate for any asset class drops below the last recorded figure.

## Capabilities

### New Capabilities
- `asset-census`: measuring and reporting how much of the installed game's content the pipeline can read and convert, per asset class, with failures grouped by cause.
- `nif-format`: which NIF block types and versions the reader supports, and that unsupported content is reported rather than silently dropped.

### Modified Capabilities

## Impact

- New command (for example `cmd/census`) and a committed report file; no game data in the repo.
- `internal/nif`: new block readers; `cmd/propconv` particles gain gravity; textures embedded in NIFs become usable.
- Conversion of all outdoor zones increases local `web/data` size (gitignored).
- Feeds `cities-and-dungeons`, whose dungeon and city pieces the census measures first.
