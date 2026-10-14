# Design

## Context

See proposal.md for why. The relevant current state:

- Reading is done by `internal/nif` (models), `internal/mpak` (archives), `internal/dds` and `internal/pcx` (images), `internal/sector` and `internal/gamedata` (tables). The NIF parser already refuses a file unless its footer lands exactly at end-of-file, so any block reader with a wrong layout fails loudly rather than producing garbage.
- Conversion logic lives in the `package main` of each command: model baking (`convertModel`, collision and shadow-hull filtering, billboards, animation, particles) in `cmd/propconv`; terrain, ground detail, sky and grass in `cmd/zoneconv`. Nothing outside those commands can call it.
- Coverage is known only from samples: `TestFiguresParse` (617 figures), `animconv -list` (38 of 4,032 `.kfa` fail), `TestEveryPlacedModelConverts` (zone 100), and an ad-hoc run of `nifdump` over the 1,343 outdoor models (1,126 parse).
- The install spreads content over `zones/` (398 zone folders, `Nifs`, `Dnifs` with 1,357 piece archives), `frontiers/` (30 zone folders, `NIFS`, `dnifs`) and `Newtowns/zones/Nifs`; `figures/` holds 633 figure files and `anims/` 4,032 animations. `zones.dat` lists 486 zones.

## Goals / Non-Goals

**Goals:**
- One command produces the whole census in minutes, using exactly the code paths the converters use.
- The committed report is stable enough to diff between commits, so a change's effect on coverage shows up in review.
- Close the seven known block types so that every file containing them parses, and give the two that change what is drawn (gravity, embedded textures) their effect.

**Non-Goals:**
- Converting city and dungeon layouts (the `cities-and-dungeons` change); the census only measures their pieces.
- Rendering `NiTextureEffect` (environment and projected textures), morph animation, alpha and material-colour animation, or UV transforms from `NiTextureTransformController`: they are read exactly and otherwise ignored, and the report notes them as read-not-drawn.
- Fixing corrupt files (3 seen so far); they are reported as such.

## Decisions

**Share conversion code through internal packages rather than shelling out.**
Move model baking from `cmd/propconv` into `internal/scenery`, and the outdoor-zone read (heightmap, fixtures resolution) into `internal/zone`; the commands become thin wrappers. The census calls the same functions in memory.
*Alternative:* run the converter binaries once per item and parse their output. Rejected: one process per model for about 4,000 models is slow, the failure causes would have to be scraped from text, and it would still not test "converts" more faithfully than calling the same code.

**`cmd/census` writes `docs/census/census.json` and `docs/census/README.md`.**
JSON is the machine-readable baseline (per class: total, passed, failures grouped by cause with counts and up to five example names, all sorted). Markdown is the human view rendered from the same data. Both are deterministic: sorted keys and lists, no timestamps, no absolute paths.
*Alternative:* print to stdout only. Rejected: the regression check and code review both need a committed baseline.

**A dry run for zones.**
For each outdoor zone the census reads the heightmap and fixtures and bakes each distinct model in memory, without writing output. Writing every zone's converted data is a separate step (`zoneconv`/`propconv` over all outdoor zones, by script), because 132 zones of output is large and gitignored anyway.

**Causes are normalised from typed errors.**
`internal/nif` returns typed errors (unsupported block type, corrupt data, footer mismatch) carrying the block type name; the census groups by type and name, never by message text with numbers in it. Errors from other readers are grouped by their type and a message with digits stripped.

**Regression guard is a Go test over the committed baseline.**
`cmd/census` has a test that runs the census and fails if any class's pass count is below `docs/census/census.json`. It is skipped with `-short` and when the install is absent. Improvements are recorded by re-running the census and committing the new report.

**Block readers follow nif.xml for versions 4.0.0.2–10.1.0.0, checked by the footer.**
Each of the seven is written from the published layout for the versions present in the install, and accepted only when every file containing it parses to end-of-file. A census run before and after shows the exact files gained.

**Gravity is analytic in the particle shader.**
Particles are drawn statelessly from the clock (see `web/particles.js`), so gravity is added as `0.5 * g * age^2` along the gravity direction transformed into model space, which matches constant acceleration exactly. Planar gravity is supported; spherical (point) gravity is approximated as planar towards the point from the emitter and noted in the report.

**Embedded textures become ordinary PNGs.**
`NiPixelData` is decoded (the palettised and RGB(A) formats present in the install) and written as `<model>_<block>.png` beside the external textures, and the model's group names it, so the viewer needs no change.

## Risks / Trade-offs

- [A block layout differs between the file versions present] → the footer check rejects it; the census lists the files still failing, and the reader branches on version where nif.xml does.
- [Census runtime grows past a few minutes] → archives are opened once and models parsed in parallel; the regression test is skipped under `-short`.
- [Report churn makes diffs noisy] → deterministic ordering and capped example lists; counts change only when coverage does.
- [Moving converter code breaks the converters] → existing converter tests (`TestEveryPlacedModelConverts`, torch and log-fire tests) must pass unchanged after the move, and zone 100's output must be byte-identical.
- [Read-but-not-drawn blocks are counted as "parsing"] → the report has a separate column for read-not-drawn features, so coverage is not overstated.

## Open Questions

- Whether to also publish the Markdown report as a page; it can be added later from the same JSON.
