# Tasks

## 1. Read the geometry that does not read yet

- [x] 1.1 Locate the non-finite field in Jordheim's `NiTriShapeData` (jord01 block 51, NaN at 57608) by logging section offsets. Read that field leniently and sanitise it as design §5 says (positions stay strict). Verify that a test converts all 31 `jordNN.nif` blocks, and that the nif tests still reject a non-finite vertex position.
- [x] 1.2 Read `NiParticleMeshes` so Trollheim's `THDHallCpool.NIF` and `THDRmfountain.NIF` convert, or, if it proves larger than a day, record them in the census allow-list with the reason. Verify with a test on both pieces, or with the allow-list entry and a README note.

## 2. Interior layouts

- [x] 2.1 Add `internal/interior` reading `dungeon.chunk`, `dungeon.place`, `dungeon.prop`, `city.csv`, `lights.csv` and `ZONEJUMP.csv` into placements (model and 4×4 transform), entrances and lights; skipped rows are counted. Verify with unit tests on zone 125 (26 pieces, 194 placements, 2 props, 82 lights, the `ExittoMyrk` entrance) and zone 120 (31 blocks).
- [ ] 2.2 Fit each interior's transform to zone coordinates by the scored candidate search in design §3, recording the transform and its score. Verify that a test prints and checks, for all 13 Midgard interiors, that at least 90% of entrance centres and lights fall inside placed piece bounds, and record each zone's fit in the README.
- [x] 2.3 Decide the dungeon prop frame per zone by the scored comparison in design §4. Verify that Abandoned Mines (707 props), Darkspire and Glashtin Forges each pick a frame with at least 90% of props inside piece bounds, and that the test fails if the frame is forced to the other reading.

## 3. Converter

- [x] 3.1 Add the optional instance `axis` to the scenery instance format (design §2), written only off ±z, with the props vertex shader honouring it. Verify zone 100's `props.json` and `props.bin` are byte-identical, and a unit test shows an axis-z instance renders the same transform as yaw alone.
- [x] 3.2 Add `cmd/interiorconv` writing `props/` and a `zone.json` with `kind`, `transform`, `entrances`, `lights` and `bounds` for an interior. Verify that zones 120 and 125 convert, the written instance counts match the layout, and the manifests list entrances and lights.

## 4. Viewer

- [x] 4.1 Move the heightmap-dependent parts of `main()` behind the `world` interface (design §7) without changing outdoor behaviour. Verify in the browser that zones 100, 102 and 152 look as before, with no WebGL error, and the node tests pass.
- [x] 4.2 Add the interior world: no terrain, water, grass or sky, dark clear colour and fog, ground from a CPU triangle grid cast downward from the character's height plus a 60-unit step, and spawn inside the first entrance. Verify in the browser that `?zone=125` opens with the character standing on the cave floor at the entrance, and that walking a sloped passage keeps the feet on the floor.
- [x] 4.3 Add point lights to the shared lighting (8 nearest, quadratic falloff, `uSun` 0 inside, no shadow pass inside). Search the install for a light-type table for `lights.csv`'s type column, else use the warm default. Verify in the browser that Spindelhalla is lit around its torches and dark between them, and that zone 100 is unchanged.
- [x] 4.4 Show Jordheim and the dungeons in the zone picker with their kind. Verify the picker lists them, marked city or dungeon, and that `[`/`]` step through them.

## 5. Census, worldconv and the guard

- [x] 5.1 Dry-run city and dungeon zones in the census (design §9): placements, failures, fitted transform and score, entrances and lights inside, with an interiors table in the README. Verify two runs are byte-identical, zone 125's row shows 196 of 196, and no install path appears.
- [x] 5.2 Make `worldconv` convert the interiors the census marks convertible, check their counts, and write `kind` to `zones.json`. Verify it reports every zone ok, including the 13 Midgard interiors.
- [x] 5.3 Add the Midgard interiors guard test (zones 120, 125–129, 149, 150, 160, 161, 188, 189, 226), skipped under `-short` or without an install. Verify it passes, and fails when one zone's transform is forced off.
- [x] 5.4 Document the interior formats, the transform fit and the lighting in the README, and update its limitations (no wall collision, no zone travel, no dungeon water). Verify the documented commands run as written.

## 6. Integration

- [ ] 6.1 Tour all 13 Midgard interiors in the browser: spawn view, a walk along a passage, and a free-fly overview, with frame rate noted for Abandoned Mines. Note any zone that still looks wrong as a finding, not a pass.
- [x] 6.2 Run `go test ./...` and the node tests. Verify all pass.

## Workflow follow-up

- Commit and push only when the user asks. After the push, verify that `git ls-tree -r --name-only origin/main | grep -ci "web/data"` returns 0.
- Archive with `/opsx:archive jordheim-and-midgard-dungeons` once reviewed.

## Closed

Archived on 2026-10-19 with the unchecked tasks above moved to `midgard-playable` (section "Carried over").
