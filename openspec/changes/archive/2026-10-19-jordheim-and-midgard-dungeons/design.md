# Design

## Context

See proposal.md for the zones and what the survey found. These facts shape the approach:

- **Scenery already bakes.** `internal/scenery` bakes any model into the instanced format `web/props.js` draws (`props.bin` and `props.json`: models, then instances with position, yaw and scale). Dungeon pieces read through it today: every piece of eleven Midgard dungeons converts, with search dirs `zones/Dnifs`, `frontiers/dnifs` and `zones/Nifs`.
- **Instances only turn about z.** An instance carries a yaw and a uniform scale. Dungeon placements carry a full axis-angle, but every survey row so far rotates about ±z, or about an axis within 0.0003 of it (`3.141593, 0.000244, 0.000244, 1`). Props may differ.
- **The viewer assumes a heightmap.** In `web/viewer.js`, `main()` derives `groundAt`, the spawn, grass, NPC picking and the camera's floor from `heights.u16` and the grid. The lighting in `web/lighting.js` is a sun with cascaded shadows plus sky ambient. There are no point lights.
- **Layout files in a dungeon's `dat<zone>.mpk`:**
  - `dungeon.chunk`: piece names, one a line.
  - `dungeon.place`: `index, x, y, z, angle, ax, ay, az, ?, ?, ?, ?, ?`.
  - `dungeon.prop`: `index, x, y, z, angle, ax, ay, az, a, b, scale`. Its index goes into the same chunk list (Abandoned Mines' largest is 48, of 55).
  - `lights.csv`: `x, y, z, type`, in zone coordinates.
  - `ZONEJUMP.csv`: `id, name, x1, y1, x2, y2, z1, z2, target`, also in zone coordinates.
- **Jordheim** has `city.csv` (a count, then `n, jordNN.nif`), `adjacency.csv` (block visibility), `lights.csv`, `ZONEJUMP.csv`, and a SECTOR.DAT with `[rotate] rotate=2`. Block bounds show the 31 blocks share one city space. 28 of the 31 stop on a non-finite float in `NiTriShapeData`, about 8 KB into a 113-vertex block of a 10.1 file. That is past the positions, normals and colours, so it is likely in an extra UV set.

## Goals / Non-Goals

**Goals:**
- Every Midgard interior is converted, opened and walkable, using the same scenery pipeline as outdoors.
- The coordinate transforms are established from the data and checked, not assumed.

**Non-Goals:**
- Wall collision, and zone transitions through the exits.
- `adjacency.csv` culling, unless frame rate demands it (Abandoned Mines, 271 pieces and 707 props, is the stress case).
- Water in dungeons.
- Light flicker.
- NPCs inside.

## Decisions

### 1. One converter for interiors, reusing scenery baking
Add `internal/interior`. It reads a zone's layout into a list of placements, each a model file with a 4×4 transform, plus entrances and lights. Add `cmd/interiorconv`, which bakes the placements through `scenery.ConvertModel` into the existing `props.bin` and `props.json` format, and writes `zone.json` with `kind`, `transform`, `entrances`, `lights` and `bounds`.

The census dry-runs through the same reader, as it does outdoors.

*Alternative:* a mode of `propconv`. propconv is built around `fixtures.csv` and nifs ids, and interiors share only the baking, so a separate small command is clearer.

### 2. Full rotations, only where the data needs them
The instance format gains an optional `axis` ([x, y, z], with `yaw` as the angle about it), written only when the axis is not ±z within 1e-3. Outdoor `props.json` is unchanged, and zone 100 stays byte-identical. The vertex shader builds the rotation from axis and angle, which for ±z is the same yaw as now.

The design does not round near-z axes to z. `0.000244, 0.000244, 1` is z to within 0.02°, and either reading draws the same.

### 3. The interior transform is fitted from the zone's own data
For each interior, `internal/interior` searches a small set of candidates, mapping layout space to zone space:
- **Translation:** the origin, from {24576, 32768, 16384, 8192} per axis.
- **Turn:** a 90° turn about z (0–3).
- **Mirror:** an optional mirror in y.

Each candidate is scored by how many entrance-box centres and lights fall inside the placed pieces' bounds, tested per piece rather than against the whole, so a candidate must put lights inside actual rooms. The best is recorded in the manifest with its score. A zone whose best leaves more than 10% of its points outside is reported, not silently converted.

The survey's hint (Spindelhalla's exit lands in the entrance piece at origin (24576, 24576, 16384)) is the first candidate, not an assumption. Jordheim's `rotate=2` is tried as the turn. The candidates are small integers, so a wrong fit stands out.

*Alternative:* hard-code 24576/24576/16384. Tuscaran Glacier's exit already sits 2,300 units outside its bounds under that origin, so it would be wrong somewhere.

### 4. Dungeon props: frame decided by the same test
`dungeon.prop`'s columns 9–10 may tie a prop to a placement, as piece-local coordinates. Two readings are scored the same way: dungeon space, and local to placement `b` (column 10). The score is the fraction of props whose position falls inside a placed piece's bounds; the winner is taken per zone and recorded. Abandoned Mines' 707 props make the comparison decisive.

### 5. Jordheim's NaN: read leniently, then sanitise
First locate the field. Bisect the `NiTriShapeData` read on `jord01` block 51, logging each section's offset against the NaN at 57608.

If the NaN is in a UV set other than the first, read UV sets leniently. Only set 0 is drawn, and non-finite UVs there become 0. If it is in normals, read them leniently and recompute non-finite ones from the faces.

Positions are never relaxed: a non-finite position still fails. The census keeps counting whichever field it was.

### 6. `NiParticleMeshes`
It is `NiParticles` whose particles are meshes. Read it with the layout nif.xml gives for 4.x/10.x. If the shape's own geometry is all that draws (the meshes being the falling water), bake the static part and let the emitter path ignore the mesh particles. Trollheim's two fountain pieces then convert. If reading it proves larger than a day, record the two pieces as known non-converting in the census allow-list and move on. That allow-list is the only exception the guard permits.

### 7. The viewer gets a world interface, outdoor or interior
`main()` is split so that what depends on the heightmap sits behind a small `world` object:
- `kind`;
- `groundAt(x, y, fromZ)`;
- `isWet(x, y)`;
- `spawn`, `bounds` and `extent`;
- `drawGround(pass)` for terrain and water, and `grass`;
- `light` setup.

The outdoor world is today's code moved, not rewritten, so zone 100 renders identically. The interior world:
- draws no terrain, water, grass or sky; the clear colour is near-black, and fog is a dark colour over a short range;
- answers `groundAt` by casting down onto the interior's triangles. The viewer keeps the props geometry on the CPU: `props.bin` is already fetched, and the instances are transformed once into a uniform grid of triangles, 512-unit cells, xy-binned. The ray starts at `fromZ` plus a step height of 60 units and takes the nearest hit below. With no hit, the last height is kept;
- spawns the character just inside the first entrance box: at the box centre, moved 200 units toward the nearest piece centre, on the ground found there, facing that piece.

Character, camera, combat and NPC code take `world` instead of reading `heights`. NPCs stay outdoor-only, as now.

*Alternative:* a separate `interior.html` and module. That duplicates the character, camera and combat setup, which is most of `main()`.

### 8. Interior lighting: baked shade, ambient, nearest point lights
`LIGHT_GLSL` gains a fixed array of 8 point lights (position, colour, radius) and a `uSun` weight:
- Outdoors, `uSun` is 1 and there are no point lights, so the output is unchanged.
- Inside, `uSun` is 0, the ambient is low, and each frame binds the 8 lights nearest the camera, with a smooth quadratic falloff.

Shadow maps are not used inside.

A light's colour and radius come from its type column if a type table turns up in the install (a task searches for it). Otherwise they default to a warm torch colour and 700 units. Models' baked vertex shading (`Vertex.C`) already darkens corners and is kept.

### 9. Census and guard
`cmd/census` lists city and dungeon zones present and dry-runs them through `internal/interior`. Each row records:
- placements and those converting;
- failing models, with causes;
- the fitted transform and its score;
- entrances and lights inside, out of their totals.

The README gains an interiors table. A Midgard interiors test mirrors the outdoor one. `worldconv` runs `interiorconv` for the interiors the census marks convertible and checks the instance counts. `zones.json` gains `kind`.

## Risks / Trade-offs

- **[Transform search finds nothing clean]** → The scores are reported per zone. A dungeon below 90% is surfaced for inspection, not converted silently.
- **[Prop frame differs per dungeon]** → It is decided per zone and recorded. A zone where neither reading scores 90% is reported.
- **[Ray-cast ground on large interiors]** Abandoned Mines is probably over a million triangles. → A grid built once at load, testing only the cell under the character; the build time is measured in the tour task.
- **[Viewer refactor regresses outdoors]** → Zone 100 screenshots before and after, and `gl.getError()` checks on 100, 102 and 152.
- **[Draw cost in big dungeons]** → Instancing as outdoors. If Abandoned Mines falls under about 20 fps, `adjacency.csv`-style culling is proposed as a follow-up, not folded in.
- **[Light types unknown]** → A default warm light. The look is checked by eye, and type-based colours come only if the table is found.

## Migration Plan

Additive. Outdoor zones' data and manifests are unchanged, apart from the new optional `kind` in `zones.json`, which defaults to outdoor. Re-run the census, then worldconv.

## Open Questions

- Whether the install carries a light-type table for `lights.csv`'s fourth column. The default colour is used either way, so this does not change the plan.
