# daoc-web

A proof of concept: extract Dark Age of Camelot's shipped assets and render
them in a browser. Currently one zone — **100, Vale of Mularn** — with a
playable **Norseman** you can walk around it.

Written in Go, with a dependency-free WebGL2 viewer.

## What it does

```
zones/zone100/dat100.mpk ──► terrain.pcx + offset.pcx ──► heights.u16 (256×256 uint16)
                         └─► SECTOR.DAT ──────────────► zone.json    (scale, water, fog)
zones/zone100/lod100.mpk ──► 64 × DXT1 tiles ─────────► atlas.png    (2048×2048)

figures/NVikingM.NIF ───────► 49 skinned shapes ──────► mesh.bin     (skinned vertices)
                         └─► 114-bone Biped skeleton ► char.json    (bones, inverse binds)
figures/skins/*.mpk ────────► DXT3/DXT5 skins ────────► tex/*.png
                                                            │
                                                            ▼
                                            web/ ── WebGL2, GPU skinning
```

## Running it

Needs Go 1.21+ and a DAoC install.

```bash
go run ./cmd/zoneconv -zone 100 -name "Vale of Mularn"   # terrain
go run ./cmd/charconv                                    # character
go run ./cmd/serve                                       # then open localhost:8777
```

Controls: `WASD` to move, drag to orbit the camera, wheel to zoom, shift to
sprint, alt to walk, `C` for the free-fly camera, `F` for wireframe. The
console exposes a `daoc` handle — `daoc.warp(60, 70)` puts the character on a
heightmap cell.

Movement uses DAoC's own rates: 85 units/s walking, 191 running, and the gait's
phase advances with ground covered rather than with time, so the feet keep pace
at any speed.

## Tooling

| command | what it does |
| --- | --- |
| `zoneconv` | zone terrain, texture atlas and water |
| `charconv` | one character: mesh, skeleton, skin weights, textures |
| `charshot` | renders an exported character to a PNG, no browser needed |
| `nifdump` | inspects a `.nif`; `-bind` checks the bind pose, `-shapes` lists geometry |
| `mpakls` | lists and extracts from MPAK archives |
| `serve` | static file server for `web/` |

`charshot` exists because a WebGL viewer cannot be inspected from a terminal,
and "the numbers look right" is a weak claim about a mesh. It reads exactly
what the browser reads, skins and poses it on the CPU and rasterises it. It
caught two bugs the numeric checks passed over.

## Verifying

```bash
go test ./...            # container, heightmap and NIF parsing against the real files
node web/skeleton.test.mjs   # the posing maths and the exported character
```

The strongest check spans both. `charconv` computes the model's bounding box by
walking the NIF's own scene graph; `skeleton.test.mjs` skins every vertex the
way the shader does, from the exported skeleton, inverse bind matrices, bone
indices and weights. The two routes share no code, and they agree to 0.08 units
on a 71.8-unit figure.

## File formats, as worked out from the shipped data

### MPAK / NPK container

```
0x00  char[4]    "MPAK"
0x04  uint8      version (observed: 2)
0x05  byte[16]   obfuscated header: real[i] = stored[i] ^ i
                   +0   uint32  checksum
                   +4   uint32  compressed size of the directory
                   +8   uint32  compressed size of the archive-name stream
                   +12  uint32  file count
0x15             zlib stream: the archive's own name
0x15+nameSize    zlib stream: directory, count × 284 bytes
then             file data, one zlib stream per entry
```

Directory entry, 284 bytes:

```
0x000  char[256]  file name, NUL-terminated
0x100  uint32     mtime (unix)        0x10C  uint32  uncompressed size
0x104  uint32     flags (always 4)    0x110  uint32  compressed offset
0x108  uint32     uncompressed offset 0x114  uint32  compressed size
                                      0x118  uint32  crc32
```

The 256-byte name field is a reused buffer: bytes after the terminator hold
leftovers from previous writes and must be ignored.

### Terrain height

Two 8-bit PCX rasters, 256×256, combined using factors from `SECTOR.DAT`:

```
height = terrain.pcx[x,y] × scalefactor  +  offset.pcx[x,y] × offsetfactor
```

For Mularn that is `×8` and `×48`, giving 2880..11848 world units. A zone spans
65536 units per side, so one height sample covers 256 units.

Confirmed against the independently-stored lake elevation: the lake surface is
at 4372, its shoreline samples land at 4488..5456 (just above the water, as
bank points should be), and 3.9% of the zone falls below the waterline.

A unit works out at roughly an inch: the Norseman model is 71.8 units tall.

### Terrain colour

`lodNN-MM.mpk` holds an 8×8 grid of 256×256 DXT1 tiles — one per sector — which
stitch into a 2048×2048 atlas for the whole zone. Index order is `lod<x>-<y>`
with no axis flips.

That orientation is not documented anywhere, so the converter derives it: it
builds all eight candidate layouts and scores each by the Pearson correlation
between per-cell brightness and terrain height, on the assumption that high
ground is snow-covered and bright. The result is unambiguous:

```
swap=false flipX=false flipY=false   r = +0.8386   ← chosen
swap=false flipX=false flipY=true    r = +0.4492
swap=true  flipX=true  flipY=true    r = +0.4073
swap=true  flipX=false flipY=false   r = +0.3896
swap=true  flipX=false flipY=true    r = +0.3222
swap=true  flipX=true  flipY=false   r = +0.3121
swap=false flipX=true  flipY=false   r = +0.2757
swap=false flipX=true  flipY=true    r = +0.1866
```

### Water

`SECTOR.DAT` declares each body with a surface height and two shore chains,
`leftNN` and `rightNN`, in heightmap cells. They are *not* one ring: `left[i]`
and `right[i]` are opposite banks at the same point along the body, so the
surface triangulates as a strip between the two chains.

### NIF models

`figures/*.NIF` are **Gamebryo 10.1.0.0**; `items/*.nif` are the older
**NetImmerse 4.2.1.0**. Both are a flat array of blocks preceded by a header,
referencing each other by index. A block carries no length, so walking the file
means knowing the exact field layout of every type present, and one wrong field
desynchronises everything after it.

That makes the format unforgiving but also self-checking: the footer has to
land precisely at end-of-file. `Parse` enforces that, which is why a clean parse
of a 686KB, 479-block model is itself the headline assertion.

Two layouts differ from the published nif.xml schema at this version, both
found by hand-decoding the bytes where parsing desynchronised:

- **No `Group ID` on `NiGeometryData`.** The schema adds it from 10.1.0.0, but
  DAoC's 10.1.0.0 files open straight with the vertex count. Reading a Group ID
  first yields 1 vertex for a 116-vertex head mesh.
- **`Skin Partition` is on `NiSkinData`, not `NiSkinInstance`.** The field moves
  between versions and is never on both; reading it in both places loses four
  bytes.

Version 10.1.0.0 also writes a zero `uint32` ahead of the header body and of
every block.

The Norseman uses just ten block types: `NiNode`, `NiTriShape`,
`NiTriShapeData`, `NiSkinInstance`, `NiSkinData`, `NiSkinPartition`,
`NiMaterialProperty`, `NiVertexColorProperty`, `NiKeyframeController` and
`NiKeyframeData`.

### Characters

A race model is not one mesh. `NVikingM.NIF` holds 49 skinned shapes: ten
heads, five torso tiers each in two cuts, several arms, legs, boots, gloves and
cloaks, and whole-body meshes for distant rendering. A character is assembled by
choosing one shape per slot — drawing all of them stacks every outfit the race
can wear on top of itself.

The skeleton's 114 joints use **3ds Max Biped names**, listed in
`figures/animnode.dat`. That is what makes the model animatable without
touching the `.kfa` files: `Bip01 L Thigh` is unambiguously the left thigh, so a
gait can be written against the names.

The bind pose gives a way to check the transform maths. For every bone of a
skinned shape:

```
worldTransform(bone) × boneSkinTransform  ==  worldTransform(shape)
```

and `NiSkinData`'s overall transform is the inverse of the shape's world
placement. Both hold: the second to 3.8e-06 across all 49 shapes, the first to
0.18 units worst case, on a fingertip ten joints from the root where float32
has accumulated down the longest chain.

### Character textures

Models carry UVs but no `NiTexturingProperty` — the game assigns skins at
runtime from equipment, so the mapping has to be reconstructed. Textures live
in `figures/skins/*.mpk` and `figures/Mskins/*.mpk` and are named by slot, which
lines up with the shape names:

```
Body1 → m1_bodym.dds      Arms1  → m1_arm1.dds     HeadA1 → nor_m_head01.dds
LBody1 → m1_tunic1.dds    Legs1  → m1_legs1.dds    Boots1 → m1_boot1.dds
```

Prefixes are race codes (`a` Avalonian, `b` Briton, `d` Dwarf, `k` Kobold,
`t` Troll, `v` Valkyn …), with `m1`/`m2`/`m3` the Midgard sets and three-letter
forms (`nor`, `tro`, `kob`) for heads and hair.

**DAoC's character skins are not alpha-masked.** The DXT5 faces carry a
smoothly varying alpha channel the engine uses for something else:
`nor_m_head01.dds` has ten opaque texels out of 65536, so reading it as opacity
erases the face. `charconv` flattens it and keeps the colour, which is a correct
skin tone.

## Animation

There is no animation data in the pipeline yet; the gait is procedural, written
against the Biped joint names. Phase advances with ground covered, and one
intensity parameter scales the same cycle from a walk to a sprint.

A joint's local frame is whatever the artist left it as, so joints are not
rotated in their own axes. A rotation is given in the body's frame and
conjugated into the joint's:

```
R_local = transpose(W) · R_body · W
```

where `W` is the joint's bind world **rotation**. Using the full bind world
transform instead leaves a translation in `R_local` and the limb flies off its
joint — which is exactly what happened, and what
`posing rotates joints in place, never translating them` now guards.

The models are authored in a T-pose, arms straight out along the lateral axis.
Swinging them fore and aft is a rotation about that same axis, which does
nothing to an arm lying along it, so the arms are brought down to the sides
first. This applies standing still too: a T-posed idle is not an idle.

## Known limitations

- **Ground-level texturing is blurry.** The LOD atlas is 2048px over 65536
  world units — about 32 units per texel. Fine from altitude, mushy up close,
  and more obvious now there is a character standing on it. Real detail needs
  the per-sector `patch*.dds` masks in `ter100.mpk` blended against the layer
  list in `textures.csv`, which is a multi-texture splatting pass this does not
  attempt.
- **No world models.** Buildings, trees and props are `.nif` meshes placed by
  `fixtures.csv` and `nifs.csv`. Mularn village is still absent. The parser now
  handles the character block set; static props need the 4.2.1.0 generation and
  its extra block types (`NiZBufferProperty`, `NiTexturingProperty`,
  `NiSourceTexture` and friends), which is a bounded extension rather than an
  unknown.
- **No recorded animation.** The 4032 `.kfa` files are unexamined, as are the
  35 `NiKeyframeController` blocks inside the model itself. The gait is
  hand-written.
- **No collision.** The character follows the heightmap and walks through
  anything else, including across the lake surface.
- **One character.** Equipment, races, genders and the `fig3/*.mpk` armour
  meshes are all reachable through the same code but not wired up.
- Heightmap edges stop at sample 255 (65280 units), 256 units short of the
  nominal zone edge. Harmless here; matters when stitching zones together.

## Licensing

The converter and viewer are original work, MIT licensed — see `LICENSE`.

The game assets they read are copyright Electronic Arts / Broadsword Online
Games. Converting them for your own use is one thing; redistributing the
converted output is not, so **no game data is in this repository**: `web/data/`
is gitignored, and you generate it by running the converters against your own
install. Without it the viewer has nothing to show and `skeleton.test.mjs` has
nothing to check.

Nothing here touches the live game servers or its network protocol.
