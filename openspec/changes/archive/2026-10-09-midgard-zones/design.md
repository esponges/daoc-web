# Design

## Context

See proposal.md (Why) for the survey. These facts from the code and the install shape the approach:

- **Terrain shader samplers.** `TERRAIN_FS` declares `uLayers` and `uMasks` (sampler2DArray) and `uSlots` (sampler2D) beside `uAtlas`. `drawTerrain` sets their units only `if (detail)`. Unset sampler uniforms are 0, and WebGL2 fails any draw where samplers of different types share a unit. So without detail, the terrain draw never happens.
- **The `tex` archive in splat.go.** splat.go reads `tex<zone>.mpk` only to score mask layouts and blend rules against the game's pre-blended tiles. The detail layers themselves come from `ter<zone>.mpk` and `zones/TerrainTex`.
- **Tiles as BMP.** Older Midgard zones store those tiles as `tex%02d-%02d.bmp`. Zone 100 stores them as `.dds`.
- **Weighted-average zones.** splat.go already scores the weighted-average rule. When that rule wins, it returns an error instead of emitting it.
- **Frontier layer textures.** These are in `frontiers/zones/TerrainTex`. splat.go only looks in `zones/TerrainTex`.
- **Grass maps.** grass.go requires `grassmap.pcx` and `densemap.pcx` to equal the heightmap grid. SI and frontier maps are 512 px against a 256 grid.
- **Tree clusters.** `zones/trees/tree_clusters.mpk` holds `tree_clusters.csv`: `name, tree, x,y,z × 10`, where zero rows are unused slots. Base trees are `.npk` archives in `zones/trees` and `zones/Nifs`.
- **Instance counting.** The census and propconv bake through `internal/scenery`, and worldconv requires propconv's instance count to equal the census's converting count.
- **Atlas layout.** zoneconv picks one of eight LOD tile layouts by correlating brightness with height. 48 of 118 zones chose a non-canonical layout. In Midgard those are 107, 151, 154 and 155, all flat or watery, where that correlation is weak.

## Goals / Non-Goals

**Goals:**
- Every Midgard outdoor zone looks like itself in the viewer.
- Fixes are generic, so other realms benefit, but they are verified on Midgard only.
- Zone 100's output stays byte-identical, except where a fix is meant to change it.

**Non-Goals:**
- Pixel-matching the game client.
- Baked shadow maps or snow maps.
- New NPCs in other zones.
- Fixing every realm's zones in this change. Their census rows will show what remains.

## Decisions

### 1. Always bind the detail samplers
At startup, create 1×1 placeholder textures: a 2D array for layers, a 2D array for masks, and a 2D texture for slots. Bind them to units 1–3 whenever the zone has no detail, and set the three sampler uniforms unconditionally.

*Alternative:* compile a second terrain program without detail. That doubles the shader to maintain, and the placeholder fix is three textures. A regression check reads `gl.getError()` after the first frame and logs it, so this class of bug is visible next time.

### 2. Reference tiles: try `.dds`, then `.bmp`
Add a small loader that reads `tex%02d-%02d.dds` and falls back to `.bmp`, decoding the BMP with the standard library (`golang.org/x/image/bmp` is not a dependency). 24-bit and 8-bit palettised uncompressed BMP covers what the game ships. First check which BMP variants occur; the decoder handles only those and errors on anything else.

*Alternative:* skip the scoring for BMP zones and assume "painted in order, as named". That gives up the check that caught the weighted-average zones, so no.

### 3. Weighted-average blends become a `blend` field
`splatOut` gains `Blend string` with `json:"blend,omitempty"`. It is empty for painted-in-order (zone 100 stays byte-identical) and `"weighted"` otherwise.

In the shader, a `uBlend` uniform switches `detail()` from `mix(c, layer, w)` to accumulating `layer*w` and `w`, then dividing. The first slot keeps weight 1, as in the scorer.

The scorer's existing threshold for choosing a rule stays. Zones whose best agreement is still low (156 at 0.64, 158 at 0.34 painted and 0.74 weighted) are emitted anyway with the agreement recorded. The census flags agreement below 0.6 as "unverified" rather than dropping the detail.

### 4. Frontier layer directory
`zone.Zone` already knows `Frontier`. splat.go builds its TerrainTex search path as `frontiers/zones/TerrainTex` then `zones/TerrainTex` for frontier zones. This mirrors `scenery.ModelDirs`.

### 5. Masks of any square size
Masks are decoded, then nearest-sampled to 128 px when their side differs. Nearest keeps channel boundaries crisp, matching how the viewer clamps mask texels per sector. Non-square or non-power-of-two masks still error.

### 6. Grass at map resolution
When the maps are k× the grid (k an integer, in practice 2), grass cells are emitted at map resolution: `cells` becomes `(grid*k)²` and the manifest gains `cellsPerSide`. In the viewer, grass.js derives a cell's world size from `zoneUnits / cellsPerSide` instead of the heightmap cell.

*Alternative:* downsample the maps to 256 by majority vote. That loses thin grass bands along paths, and the viewer change is small.

The 103 warnings ("grassmap values with no group") are a separate symptom: values that aren't multiples of 10. Investigate them in a task, reading the CSV's group column against the map's palette, before deciding.

### 7. Tree clusters expand inside `internal/scenery`
`scenery` loads `tree_clusters.csv` once per game directory. When `FindModel` is asked for a file named in the table (case-insensitive), the scenery package handles it in three steps:
1. It bakes the base tree, searching `zones/trees` after the usual dirs.
2. It concatenates that geometry once per non-zero offset, translating each copy by (x, y, z) in model space before the placement transform.
3. It returns the result as one model named after the cluster.

A cluster then stays one placement and one instance, so propconv's instance count still equals the census's converting count. Texture references are the base tree's, so they dedupe.

*Alternative:* emit one instance per tree. That is closer to the game's culling, but it breaks the placement-equals-instance invariant that worldconv checks, and gains nothing at our draw distances.

*Offset axes:* the CSV z is 0 throughout, so offsets are in the model's ground plane. A test bakes `NPineACL5` and checks the copies' bounds centres equal the listed x and y, in the same axis convention the placed models use.

### 8. Atlas layout: canonical, checked against reference tiles
The LOD tile names are a file format, not a per-zone choice. zoneconv uses `swap=false flipX=false flipY=false`, which every strongly correlating Midgard zone chose, unless it is contradicted. As the check, correlate the stitched atlas, downsampled, against the zone's `tex` reference tiles. Those are the same sectors at a different resolution, a far stronger signal than brightness against height. Report r for the canonical layout and for the best other layout.
- **Contradicted:** if another layout beats canonical by a clear margin, keep that layout and print a warning.
- **No reference tiles:** where `tex` tiles are absent, fall back to the brightness heuristic.

*Alternative:* trust the heuristic and only add the check. The check is what tells us whether the four Midgard outliers are wrong, and the decision follows from it. The task verifies which layout the tiles support before changing any default.

### 9. Census fidelity fields
`ZoneRow` gains three fields, each "ok" or the reason, from the same code paths zoneconv uses:
- `detail`: computed by calling splat's check in dry-run mode.
- `grass`
- `layout`: the layout chosen and its r.

To make that possible, move the splat and grass readers from `cmd/zoneconv` into `internal/zone` (or a new `internal/ground`) so the census can call them without writing files, the same move world-census made for scenery.

The Midgard test reads the census rows for the 25 zone numbers. It requires terrain "ok", detail "ok", grass "ok" or "no grass map", and `converting == placements`, except for models listed in an explicit allow-list, each with a reason.

## Risks / Trade-offs

- **[BMP variants]** The tex BMPs may be a variant the small decoder does not handle. → Survey them first, and support exactly what ships.
- **[Weighted blend in the shader]** Dividing by the summed weight where all masks are 0 except the base is fine, because the base has weight 1. Precision at mask edges may show seams. → Compare against the reference tile in the browser for 108.
- **[Low-agreement SI zones (156, 158)]** Their detail may look wrong, not just slightly off. → They are emitted with agreement recorded. If a browser check shows clearly wrong ground, drop detail for that zone, with the reason recorded in the census.
- **[Moving splat and grass into an internal package]** This is a refactor on the way. → Zone 100's output must stay byte-identical across the move (checked against `base.md5`), as with world-census.
- **[Cluster geometry size]** A 10-tree cluster is 10× the base triangles in one model. → Fine at current counts. propconv already reports triangle totals, so watch them.
- **[Ncarcass `NiLODNode` corruption]** It may be a real format variant (LOD ranges with a 10.x extra field), or junk. → Inspect it. If it is junk, record it in the allow-list rather than special-casing it.
- **[Re-conversion time]** worldconv over 118 zones takes several minutes. → Iterate with zoneconv and propconv on single zones, and run worldconv at the end.

## Migration Plan

All output lives in `web/data`, which is not committed. After implementing, run `go run ./cmd/census` and then `go run ./cmd/worldconv`. Old `zone.json` files without `blend` or `cellsPerSide` still load: both default to the current behaviour.

## Open Questions

- Do the "no drawable geometry" huts (`nrs-hut1_normal.nif`, `nrshse1_normal.nif`, `nrs-hut1_frozen.nif`) draw in the game at all, or are they collision-only? To be answered by inspecting their blocks. Either outcome fits the allow-list mechanism.
