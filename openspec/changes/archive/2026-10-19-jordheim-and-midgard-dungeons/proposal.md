# Proposal

## Why

Midgard's outdoor zones now draw completely, but its capital and its dungeons cannot be opened at all. Neither has a heightmap: they are assembled from model pieces, and no converter reads their layouts.

Jordheim (zone 120) and twelve Midgard dungeons are in this install:
- **Classic:** Spindelhalla, Vendo Caverns, Varulvhamn, Cursed Tomb and Nisse's Lair.
- **Shrouded Isles:** Nyttheim, Trollheim, Tuscaran Glacier and Iarnvidiur's Lair.
- **Catacombs:** Darkspire, Glashtin Forges and Abandoned Mines.

Bringing them in finishes Midgard as a realm you can walk through, and it builds the interior path the other realms' cities and dungeons will reuse.

A survey of the files shows the layouts are simple, and most of the geometry already reads:

- **Dungeons:**
  - `dungeon.chunk` lists piece models.
  - `dungeon.place` places them by index with a position and an axis-angle rotation.
  - `dungeon.prop` places props by the same index list.
  - Every piece in eleven of the twelve dungeons converts today. Trollheim's two fountain pieces stop on an unread `NiParticleMeshes` block.
- **Jordheim:** `city.csv` lists 31 blocks, `jord01.nif` to `jord31.nif`. Each is already positioned in city space, so no placement is needed. 28 of the 31 stop on a NaN inside their mesh data, apparently in extra UV channels.
- **Coordinates:** exits (`ZONEJUMP.csv`) and lights (`lights.csv`) are in zone coordinates, while pieces are placed around a local origin. The transform between the two, apparently an origin near (24576, 24576, 16384) for dungeons, is not documented. It has to be established from the data.

## What Changes

**Conversion:**
- A converter for interior zones writes Jordheim and each Midgard dungeon into `web/data` in the scenery format the viewer already draws. It records each zone's kind, its transform to zone coordinates, its entrances and its lights.
- The NIF reader tolerates the non-finite values in Jordheim's blocks, so all 31 draw. It also reads `NiParticleMeshes`, so Trollheim's fountains convert.
- Dungeon props are placed with their pieces, by whatever frame the data shows they use.

**Viewer:**
- The viewer opens an interior zone from the same `?zone=` parameter and zone picker.
  - Interiors get no terrain, water, grass or sky.
  - The character starts at an entrance from `ZONEJUMP.csv`.
  - The ground under the character comes from the geometry below it.
- Interiors are lit by the models' baked vertex shading, an ambient level and the zone's point lights from `lights.csv`. The outdoor sun and its shadows are not used inside.
- `worldconv` converts the Midgard interiors as well, and `zones.json` marks each zone's kind.

**Census:** the census dry-runs city and dungeon zones as it does outdoor ones. A test holds Jordheim and the twelve Midgard dungeons complete.

**Out of scope:**
- Walls that stop the character (collision), and zone-to-zone travel through the exits.
- Other realms' cities and dungeons (the `cities-and-dungeons` change).
- The Atlantis and Darkness Falls zones.
- Dungeon water surfaces (`dungeon.water`).
- `adjacency.csv` visibility culling, unless performance needs it.

## Capabilities

### New Capabilities
- `interior-layout`: reading a city or dungeon zone's layout (pieces, placements, props, lights, entrances) into baked geometry in zone coordinates.
- `interior-viewing`: the viewer presenting an interior zone, covering how it is entered, grounded and lit.

### Modified Capabilities
- `zone-fidelity`: the census and its guard extend from outdoor zones to Midgard's city and dungeons (added requirements; the outdoor ones are unchanged).

## Impact

- **New code:**
  - an interior layout reader in `internal/zone` or a new `internal/interior`;
  - a converter, either `cmd/interiorconv` or a mode of `propconv`.
- **Changed code:**
  - `internal/nif`: lenient UV/normal reads and `NiParticleMeshes`;
  - `cmd/census` and `cmd/worldconv`;
  - `web/viewer.js`: an interior path that leaves out terrain, water, grass and sky, with ground found by ray, point lights and an entrance spawn;
  - `web/props.js` and `web/lighting.js` for point lights.
- **Output:** `web/data` grows by the interiors' scenery, which stays uncommitted.
- **Docs:** `README.md` (interior formats, limitations), regenerated `docs/census/*`.
