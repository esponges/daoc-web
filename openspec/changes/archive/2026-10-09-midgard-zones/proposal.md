# Proposal

## Why

worldconv converted 118 zones, and the census reports their terrain readable and 95% of their props converting. Most of them still look broken in the viewer: water and trees float over a flat white or grey void.

A survey of the 25 Midgard outdoor zones in this install found the causes:

- **No terrain in 17 zones.** Every zone that zoneconv wrote without ground detail draws no terrain at all. The terrain shader's array samplers default to texture unit 0, where the atlas is bound as a 2D texture. WebGL rejects that draw call, so only the clear colour shows. Zone 100, the one zone tuned by hand, always had detail, which is why this never showed up.
- **No ground detail in 16 zones.** Four separate gaps in zoneconv:
  - Older zones keep their reference tiles as `tex*.bmp`, not `.dds` (102–107, 111–113).
  - Some zones blend their layers as a weighted average, which the viewer does not draw (108, 153, 154, 158).
  - Frontier layer textures live in `frontiers/zones/TerrainTex` (167–169).
  - Some masks are 64 px rather than 128 px (170).
- **Frontier trees missing.** The four Midgard frontier zones lose 35–67% of their props. Their tree "clusters" (`*CL5.nif`, `*CL10.nif`) are not model files: `zones/trees/tree_clusters.mpk` lists each one as a base tree repeated at up to ten offsets.
- **No grass in 11 zones.** Grass is skipped in every Shrouded Isles and frontier zone because their grass maps are 512 px against a 256-cell heightmap.
- **Smaller gaps:** 98 placements fail on two models (a corrupt `NiLODNode` in `Ncarcass.nif`, and huts with no drawable geometry). In three zones the atlas alignment is unverified (106, 107, 111).

The census measured whether files *read*, not whether a zone *looks right*. That gap hid all of this. Midgard comes first because the starting zone is Midgard. Most of the fixes are generic, so the other realms will benefit too.

## What Changes

**Viewer:**
- Terrain draws in every zone, with or without ground detail.
- The viewer can draw weighted-average layer blends.

**zoneconv ground detail:**
- Reads `.bmp` reference tiles.
- Searches the frontier `TerrainTex` directory.
- Resamples masks that are not 128 px.
- Emits weighted-average blends instead of skipping them.

**Grass:** zoneconv reads 512 px grass and density maps against a 256-cell heightmap.

**Tree clusters:** the scenery package expands each tree cluster into its base tree at each listed offset, so frontier forests appear.

**Model failures:** the two remaining Midgard model failures are fixed, or each is recorded with a verified reason it cannot draw.

**Alignment:** in the three weak-correlation zones, the atlas alignment is checked against something stronger than brightness against height.

**Census:**
- Each outdoor zone's row records how it will *look*: terrain drawn, ground detail present (or why not), grass present (or why not), and props converting.
- A Midgard fidelity test fails if any Midgard zone regresses on any of these.

**Out of scope:**
- Jordheim (a city, under cities-and-dungeons).
- Midgard dungeons.
- The housing zones and Atlantis zones, which are not in this install.
- Shadow maps (`shadow.pcx`) and snow maps (`snowmap.tga`).

## Capabilities

### New Capabilities
- `zone-fidelity`: what a converted outdoor zone must show in the viewer (terrain, ground detail, grass, placed scenery including tree clusters), and how the census reports and guards that per zone.

### Modified Capabilities
(none: `asset-census` and `nif-format` are still unarchived in world-census. This change adds the per-zone visual columns as a new capability and does not edit those requirements.)

## Impact

- **Viewer:**
  - `web/viewer.js`: terrain shader sampler units, and a weighted-average blend path.
- **zoneconv:**
  - `cmd/zoneconv/splat.go`: `.bmp` reference tiles, frontier `TerrainTex`, mask resampling, emitting weighted-average blends.
  - `cmd/zoneconv/grass.go`: 512 px maps.
- **Scenery and NIF:**
  - `internal/scenery`: expanding tree clusters, used by both propconv and the census.
  - `internal/nif`: whatever the `NiLODNode` failure turns out to need.
- **Census:**
  - `cmd/census`: new per-zone fidelity fields and the Midgard test.
  - `docs/census/*`: regenerated.
- **README:** limitations updated.
- **Output:** re-running worldconv rewrites `web/data`, which stays uncommitted.
