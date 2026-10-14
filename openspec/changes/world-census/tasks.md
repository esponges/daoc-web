# Tasks

## 1. Share conversion code

- [x] 1.1 Move model baking (convertModel and its helpers: collision and shadow-hull filtering, billboards, UV scroll and pulse, particles, texture extraction) from `cmd/propconv` into `internal/scenery`, leaving `propconv` as a wrapper; verify `go test ./...` passes unchanged and `propconv -zone 100` writes a byte-identical `props.bin` and `props.json`
- [x] 1.2 Move the outdoor-zone read (SECTOR.DAT, heightmap, fixtures and nifs.csv resolution) into `internal/zone`, used by `zoneconv` and `propconv`; verify `zoneconv -zone 100` output is byte-identical and the zoneconv tests pass
- [x] 1.3 Give `internal/nif` typed errors (unsupported block type with its name, corrupt data, footer mismatch) and keep current messages; verify with a test that an unsupported-block file yields the unsupported-type error naming the block, and a non-finite-float file yields the corrupt-data error

## 2. Census command and report

- [x] 2.1 Add `cmd/census` that enumerates zones from `zones.dat` by type (marking missing folders) and every asset class across `zones/`, `frontiers/` and `Newtowns/`: scenery models, dungeon pieces, city block models, figures, animations, referenced textures; verify per-class totals printed for the local install match `ls` counts of those folders
- [x] 2.2 Read each item through the shared readers and record pass or fail, continuing past errors and panics; verify a run completes over the whole install and reports 1,126 of 1,343 `zones/Nifs` models passing, matching the earlier `nifdump` measurement
- [x] 2.3 Group failures by normalised cause with counts and up to five example names, ordered by count; verify the seven known block types appear with counts 70, 64, 37, 24, 7, 6 and 5 for `zones/Nifs`
- [x] 2.4 Dry-run every outdoor zone (terrain plus scenery bake in memory) and record per-zone placements converting; verify zone 100 reports 1,006 of 1,006
- [x] 2.5 Write deterministic `docs/census/census.json` and `docs/census/README.md` (sorted, no timestamps or absolute paths, no game content); verify two consecutive runs produce identical files and a grep finds no install path in them
- [x] 2.6 Add the regression test comparing pass counts with the committed `census.json`, skipped under `-short` and without an install; verify it passes on the current baseline and fails when a baseline count is raised by hand
- [x] 2.7 Document the census (how to run it, what each class means, how to update the baseline) in the README's Verifying section; verify the documented command runs as written

## 3. Close the known NIF block types

- [x] 3.1 Read `NiTextureEffect`, `NiTextureTransformController`, `NiAlphaController` and `NiMaterialColorController` exactly; verify every file the census listed under them now parses to end-of-file and the census shows zero failures for those causes
- [x] 3.2 Read `NiGeomMorpherController` and its morph data exactly; verify the 24 files listed under it parse and their base geometry converts
- [x] 3.3 Read `NiGravity` (force, type, position, direction) and carry it into the particle settings `internal/scenery` writes; verify the 64 files parse and a test checks a known model's gravity force and direction
- [x] 3.4 Apply gravity in `web/particles.js` as constant acceleration in model space; verify in the browser on a gravity emitter that particle paths curve, and that the log fire, which has no gravity, is unchanged
- [x] 3.5 Read `NiPixelData`, decode its formats to PNG and use it as the model's texture when the source texture is internal; verify the 7 files parse, a test decodes one to the expected size, and the model draws textured in the viewer
- [x] 3.6 Add `read-not-drawn` tracking so the report lists features parsed but not rendered (texture effects, morphs, alpha and colour animation, texture transforms); verify the README report has that section with counts
- [x] 3.7 Document the new block types and what is and is not drawn in the README's NIF and Known limitations sections; verify each listed block type matches the reader

## 4. Whole-world baseline

- [x] 4.1 Re-run the census and commit the report as the new baseline; verify outdoor scenery parse rate is at least 99% and the remaining failures are named in the report
- [x] 4.2 Add a script that converts every outdoor zone the census marks convertible with `zoneconv` and `propconv`; verify it completes and each converted zone's `zone.json` and `props.json` exist with placement counts matching the census (viewing them needs the zone picker from `cities-and-dungeons`)
- [x] 4.3 Verify on the remote that no game data was committed (`git ls-tree -r --name-only origin/main | grep -ci web/data` returns 0)
