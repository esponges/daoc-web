# Tasks

## 1. Terrain draws without ground detail

- [x] 1.1 Create 1×1 placeholder textures for the layer array, mask array and slot table, and bind them on units 1–3 with the sampler uniforms set unconditionally in `drawTerrain`. Verify in the browser:
  - Zone 102 (West Svealand) shows lit, atlas-coloured terrain in the free-fly overview.
  - `gl.getError()` is 0 after a frame.
  - Zone 100 looks unchanged.
- [x] 1.2 Log any `gl.getError()` after the first frame as a console warning naming the zone. Verify that forcing the old unbound state by hand produces the warning, and that the current build produces none on zones 100, 102 and 152.
- [x] 1.3 Treat placement scales above 50× as table typos and use 1. Verify that a zone test pins zone 102's 175150% oak to 1× and that zone 102 renders to the horizon (the oak walled off half the zone).
- [x] 1.4 Release the WebGL context on `pagehide`, since moving between zones reloads the page. Verify that twelve zone loads in one tab leave it responsive; before this, the GPU stalled after about ten.

## 2. Ground detail for every Midgard zone

- [x] 2.1 Move the splat and grass readers from `cmd/zoneconv` into an internal package that the census can call without writing files. Verify `zoneconv -zone 100` output is byte-identical to `base.md5` and the zoneconv tests pass.
- [x] 2.2 Survey the `tex*.bmp` variants in Midgard's tex archives (bit depth, palette, compression), and add a reference-tile loader that tries `.dds` then `.bmp`, supporting exactly the variants found. Verify:
  - A unit test decodes one BMP tile from zone 102.
  - `zoneconv -zone 102` prints a mask-layout agreement instead of "not found".
- [x] 2.3 Search `frontiers/zones/TerrainTex` before `zones/TerrainTex` for frontier zones. Verify zoneconv on zones 167, 168 and 169 finds `Odin_color`, `Jamtland_color` and `Yggdra_color`.
- [x] 2.4 Resample square masks of any size to 128 px with nearest sampling. Verify:
  - A unit test turns a 64 px mask into a 128 px one with crisp channel edges.
  - `zoneconv -zone 170` writes ground detail.
- [x] 2.5 Emit weighted-average blends as `"blend": "weighted"` (omitted when painted), and draw them in the shader by normalising by the summed weights. Verify:
  - Zone 100's zone.json is byte-identical.
  - Zones 108, 153, 154 and 158 get ground detail.
  - In the browser, zone 108's near ground matches its atlas colours as the detail fades out with distance.
- [x] 2.6 Record each zone's blend agreement. Then browser-check zones 156 and 158 (low agreement): if the ground is clearly wrong, drop that zone's detail and record the reason. Note the outcome per zone in the census README. Outcome: no zone was clearly wrong up close (105 at 0.56, 111 at 0.66, 156 at 0.64, 158 at 0.74), so all keep their detail. A margin rule now keeps the named mask layout unless another wins by 0.1; that settled 103, 107 and 158, whose near-tie layouts had been picked.
- [x] 2.7 Update the README's terrain section (BMP reference tiles, frontier layers, weighted blends). Verify that a zoneconv run over all 25 Midgard zones prints no "ground detail skipped", apart from any reasons recorded in 2.6.

## 3. Atlas layout

- [x] 3.1 Score the eight LOD layouts against the zone's reference tiles (downsampled atlas against tex tiles) where they exist, keep the brightness heuristic as a fallback, and print r for the chosen layout and the best other layout. Verify: Outcome: the reference tiles pick the canonical layout in every Midgard zone (zone 100 at r = 1.00). The heuristic had laid out 107, 151, 154 and 155 wrongly.
  - Zone 100 still chooses the canonical layout, with a high r.
  - Zones 107, 151, 154 and 155 print the layout the tiles support.
- [x] 3.2 Browser-check those four zones: shorelines in the atlas meet the water polygons and the hills. Verify with a free-fly screenshot of each, kept in the scratchpad.

## 4. Grass

- [x] 4.1 Accept grass and density maps at an integer multiple k of the grid: emit cells at map resolution and add `cellsPerSide` to the grass manifest, with the old manifests defaulting to the grid. Verify:
  - The grass test covers k=2.
  - `zoneconv -zone 152` writes grass.
  - Zone 100's grass output is byte-identical.
- [x] 4.2 Make grass.js size its cells from `zoneUnits / cellsPerSide`. Verify in the browser that zone 152 shows grass under the character and zone 100's grass is unchanged.
- [x] 4.3 Investigate zone 103's "grassmap values with no group" by comparing the map's palette values against `grass_mid.csv` groups, then fix the mapping or record why the values are unused. Verify zoneconv on zone 103 prints no warning, or the census records the reason. Outcome: a group owns its band of ten (value / 10, rounded down). Classic maps paint the foot of each band, with soft edges between (Gotar); the Isles and frontiers paint near the middle (5, 45, 105, 155). Every Midgard value falls in a group its table has, so the warning is gone.
- [x] 4.4 Report a missing grass map as "no grass map", distinct from failures. Verify zones 104, 111, 112 and 115 report that reason and nothing else.

## 5. Placed scenery

- [x] 5.1 Load `zones/trees/tree_clusters.csv` in `internal/scenery`, add `zones/trees` to the model search, and bake a cluster as its base tree copied at each non-zero offset. Verify:
  - A test bakes `NPineACL5.nif` into 5 copies whose bounds centres match the listed x and y.
  - A cluster row with zero slots yields only its listed trees.
- [x] 5.2 Run the census and propconv on zones 167–170. Verify each zone's converting count rises to its placement count, or names only non-cluster models, and the frontier forests show in the browser.
- [x] 5.3 Inspect `Ncarcass.nif`'s `NiLODNode` failure and fix the reader if it is a format variant. Verify zones 105, 106 and 116 convert it, or the census allow-list records it as corrupt with the offending offset. Outcome: a format quirk, not corruption. The file uses NaN for open ends ([NaN, 100000] and [0, NaN]), so LOD ranges are now read leniently; nothing picks a level by range.
- [x] 5.4 Inspect the "no drawable geometry" huts (`nrs-hut1_normal.nif`, `nrshse1_normal.nif`, `nrs-hut1_frozen.nif`), and either draw what they hold or record them in the allow-list with the reason. Verify the census for zones 101 and 116. Outcome: zones/Nifs keeps these as empty stand-ins, and the buildings are in Newtowns/zones/Nifs. A model that bakes empty is now looked up in the next directory. Classic zones also search frontiers/NIFS, which holds Malmohus's two ruins.

## 6. Census fidelity and the Midgard guard

- [x] 6.1 Add `layout`, `detail` and `grass` to each outdoor zone row, computed by the shared readers in dry-run mode, and add them to README.md's zone table. Verify that two census runs are byte-identical and zone 100's row reads ok, ok, ok.
- [x] 6.2 Add the Midgard fidelity test over zones 100–108, 111–113, 115, 116, 151–156, 158 and 167–170. It requires terrain ok, detail ok, grass ok or "no grass map", and every placement converting except models in an explicit allow-list with reasons; it is skipped under `-short` or without an install. Verify it passes, and fails when one zone's detail is forced to a skip reason.
- [x] 6.3 Regenerate `docs/census/*` and update the README's limitations. Verify that `git diff` shows only the expected count changes and that no install path appears.

## 7. Integration

- [x] 7.1 Re-run `go run ./cmd/worldconv`. Verify it reports 118/118 ok and zones.json is rewritten.
- [x] 7.2 Browser-tour all 25 Midgard zones: free-fly overview plus ground level at the spawn, for each. Verify every zone shows terrain, ground detail up close, grass where it has a map, and its forests. Note any zone that still looks wrong as a finding rather than a pass. Outcome: all 25 pass. One cosmetic finding: zones other than Mularn spawn at the map centre, which in Odin's Gate (115) is inside a tree.
- [x] 7.3 Run `go test ./...` and the node tests. Verify all pass.

## Workflow follow-up

- Commit and push only when the user asks. After the push, verify that `git ls-tree -r --name-only origin/main | grep -ci "web/data"` returns 0.
- Archive with `/opsx:archive midgard-zones` once reviewed.
