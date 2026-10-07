# daoc-web

A proof of concept: extract Dark Age of Camelot's shipped assets and render
them in a browser. Currently one zone — **100, Vale of Mularn**, its forest and
the village of Mularn — with a playable character you can walk around it: a
**Troll in full plate** by default, or the **Norseman** via `?char=norseman`.

Written in Go, with a dependency-free WebGL2 viewer.

## What it does

```
zones/zone100/dat100.mpk ──► terrain.pcx + offset.pcx ──► heights.u16 (256×256 uint16)
                         └─► SECTOR.DAT ──────────────► zone.json    (scale, water, fog)
zones/zone100/lod100.mpk ──► 64 × DXT1 tiles ─────────► atlas.png    (2048×2048)
                         └─► nifs.csv + fixtures.csv ─┐
zones/Nifs/*.npk ───────────► 60 scenery models ──────┴► props.bin   (baked meshes)
                                                      └► props.json  (1006 placements)

figures/NTrollM.NIF ────────► 49 skinned shapes ──────► mesh.bin     (skinned vertices)
                         └─► 114-bone Biped skeleton ► char.json    (bones, inverse binds)
figures/skins/*.mpk ────────► DXT3/DXT5 skins ────────► tex/*.png
anims/Troll_*.kfa ──────────► keyframe tracks ────────► anim/*.json  (idle/walk/run)
  + figures/animnode.dat ───► bone index -> name ──────┘
                                                            │
                                                            ▼
                                            web/ ── WebGL2, GPU skinning
```

## Running it

Needs Go 1.21+ and a DAoC install.

```bash
go run ./cmd/zoneconv -zone 100 -name "Vale of Mularn"   # terrain
go run ./cmd/propconv -zone 100                          # trees, buildings, props

# A character is an outfit: one shape per slot plus the texture to dress it.
go run ./cmd/charconv -outfit troll-plate
go run ./cmd/animconv -char web/data/char/troll-plate \
  -anims "Troll_idle,Troll_walk,troll_run" -as "idle,walk,run"

# The Norseman is still there; ?char=norseman in the URL switches to him.
go run ./cmd/charconv -outfit norseman
go run ./cmd/animconv -char web/data/char/norseman \
  -anims "I_VM,W_VM,R_VM" -as "idle,walk,run"

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
| `propconv` | zone scenery: models flattened, placed and grouped by texture |
| `charconv` | one character: mesh, skeleton, skin weights, textures |
| `animconv` | recorded animations; `-list` scores every clip against a skeleton |
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
node web/skeleton.test.mjs              # the posing maths, for the Norseman
node web/skeleton.test.mjs troll-plate  # and for any other converted outfit
```

Two checks are worth more than the rest.

**The character, two ways.** `charconv` computes the model's bounding box by
walking the NIF's own scene graph; `skeleton.test.mjs` skins every vertex the
way the shader does, from the exported skeleton, inverse bind matrices, bone
indices and weights. The two routes share no code, and they agree to 0.08 units
on a 71.8-unit figure.

**The terrain, against the game's own designers.** `fixtures.csv` gives an
absolute world height for every prop in the zone, authored by EA against EA's
terrain. This project's heightmap is derived independently, from two 8-bit PCX
rasters combined by factors read out of `SECTOR.DAT`, with the tile orientation
inferred statistically. Nothing connects the two. Of 1006 fixtures, **87.4% sit
on the derived terrain exactly** — within half a unit, where a unit is about an
inch — and the mean disagreement is **1.34 units**. A wrong scale factor, a
wrong offset or a flipped axis would put that in the hundreds.

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

DAoC ships four generations: **NetImmerse 4.1.0.12**, **4.2.1.0** and
**4.2.2.0** for the older scenery and items, and **Gamebryo 10.1.0.0** for
characters and newer props. Vale of Mularn's 81 models alone span all of them.
All are a flat array of blocks preceded by a header,
referencing each other by index. A block carries no length, so walking the file
means knowing the exact field layout of every type present, and one wrong field
desynchronises everything after it.

That makes the format unforgiving but also self-checking: the footer has to
land precisely at end-of-file. `Parse` enforces that, which is why a clean parse
of a 686KB, 479-block model is itself the headline assertion.

Eight layouts differ from the published nif.xml schema, every one of them found
the same way: parsing desynchronised, and the bytes at that offset were decoded
by hand until they said something sensible.

- **No `Group ID` on `NiGeometryData`.** The schema adds it from 10.1.0.0, but
  DAoC's 10.1.0.0 files open straight with the vertex count. Reading a Group ID
  first yields 1 vertex for a 116-vertex head mesh.
- **`Skin Partition` is on `NiSkinData`, not `NiSkinInstance`.** The field moves
  between versions and is never on both; reading it in both places loses four
  bytes.
- **The UV-set count moves.** From 10.0.1.0 it sits between the vertices and
  the normals. At 4.2.1.0 it comes after the vertex colours instead, directly
  ahead of the arrays it describes. Read in the modern position it is 22273,
  and the byte after it is the first of a unit-length normal.
- **`NiSwitchNode`'s active-child index is present at every version**, not just
  from 10.1.0.0. The pine's LOD node has two children and two ranges, 0–6575.7
  and 6575.7–100000, and they only line up if the centre starts four bytes
  further on.
- **No `Direct Render` on `NiSourceTexture`.** The schema adds it at 10.1.0.0.
  In `NF_b-fence1.NIF` the block holding `A_oldwood.dds` ends one byte before
  the flag would, and reading it eats the next block's separator.
- **`NiParticlesData` has no triangle count.** It extends `NiGeometryData`
  directly, where the two triangle blocks reach it through
  `NiTriBasedGeomData`. Reading a count there is two bytes the file never
  wrote, which is exactly how far `logfire.nif` ran over.
- **`NiDynamicEffect` carries no affected-node list** at 4.1/4.2. The schema
  offers two alternative arrays gated at 4.0.0.2 from either side, so both read
  as present; neither is. In `spikes.NIF` the `__MAX_Default_Light` block puts
  1.0f where a count would be, and that float is the dimmer.
- **`NiCamera` gates two fields the other way.** No orthographic-projection
  flag before 10.1.0.0, but a screen-texture count at every version. The
  viewport settles it: in `HCelttreetower.NIF` those four floats read 0, 1, 1,
  0 and the frustum 1.0 to 5000.0 only under that reading.

Version 10.1.0.0 also writes a zero `uint32` ahead of the header body and of
every block.

The Norseman uses just ten block types; reading the zone's scenery takes
**thirty-two**. Only `NiTriShapeData` and `NiTriStripsData` carry geometry that
is drawn. Everything else — render state, texture references, particle systems,
animation controllers, lights, a camera — is read purely to get the byte count
right, because a block carries no length and a type that merely exists in a
file cannot be stepped over. A campfire's flame is an emitter nothing here
draws, but until that emitter can be consumed exactly, the logs underneath it
cannot be read either.

Strips are converted to triangle lists at parse time — alternating winding,
degenerate stitches dropped — so nothing downstream has to know which topology
an artist used.

One deliberate hole in the checking: `Parse` rejects non-finite floats, which
is how two of the layout bugs above announced themselves, but that check is
disabled inside particle data. `NiAutoNormalParticles` means the engine
generates the normals at runtime, so the array the exporter wrote is never read
by the game, and in `NDwrfville.nif` it contains a NaN among other
uninitialised bytes. The footer landing at end-of-file still covers the block.

### Scenery

Two CSVs inside `datNNN.mpk` place the world. `nifs.csv` maps a numeric id to a
model file in `zones/Nifs`, one `.npk` archive per model; `fixtures.csv` places
that id:

```
ID,NIF #,Textual Name,X,Y,Z,A,Scale,...,3D Angle,3D Axis X,3D Axis Y,3D Axis Z
1,403,Pine,52736.00,19200.00,4960.00,225,150,...,2.356194,0.000000,0.000000,1.000000
```

Three things about that row are not obvious:

- **`Z` is absolute world height, not an offset.** It matches the heightmap
  this project derives from `terrain.pcx` to the unit, which is what the
  verification section leans on.
- **`A` and the axis-angle columns disagree, and both are right.** `A` is a
  clockwise heading in degrees; the 3D angle is the same rotation measured
  anticlockwise, with the sign carried by the Z axis column. Here 225° and
  2.356194 rad (135°) describe one heading. `propconv` takes the axis-angle
  pair, which needs no convention guessed.
- **`Scale` is a percentage**, bounded per model by the `MinScale`/`MaxScale`
  columns of `nifs.csv`.

Zone 100 places **1006 fixtures from 60 distinct models**, all of which convert,
and it is a lopsided distribution: 551 are one pine and 228 one evergreen, so
two models account for 77% of the zone. Props never move, so `propconv` bakes
each model's scene graph into flat geometry at conversion time and the viewer
draws each model once with instancing — 257 calls for the whole zone.

Models carry invisible collision hulls beside the visible mesh; the fences make
the convention clearest, with a `Collisionswitch` node branching into `collidee`
and `visible`. Those branches are dropped. `NiLODNode` subtrees keep only the
nearest level, since props are seen from the ground.

### Characters

A race model is not one mesh. `NVikingM.NIF` and `NTrollM.NIF` each hold 49
skinned shapes: ten heads, five torso tiers in two cuts apiece, several arms,
legs, boots, gloves and cloaks, and whole-body meshes for distant rendering. A
character is assembled by choosing one shape per slot — drawing all of them
stacks every outfit the race can wear on top of itself.

**The tiers are the game's armour categories, in order.** `Body1` through
`Body5` are cloth, leather, studded, chain and plate, and the same ordering runs
through `Boots1..5`, `Legs1..4` and `Gloves1..3` — a slot's last tier is its
plate variant. That is why an outfit has to pick shape and texture together: the
tier gives the silhouette and the texture gives the material, and choosing them
independently gets you plate boots painted like cloth.

`charconv -outfit` names a pairing. Two ship:

| outfit | model | shapes |
| --- | --- | --- |
| `norseman` | `NVikingM.NIF` | `HeadA1,Body1,LBody1,Arms1,Legs1,Gloves1,Boots1` |
| `troll-plate` | `NTrollM.NIF` | `HeadA1,Body5,Lbody4,Arms6,Legs4,Gloves3,Boots5` |

`-shapes` and `-tex Slot=file.dds` override either half for experiments.

Leaving a slot out is visible rather than subtle: the first troll build omitted
`Lbody4` and came out with a hole through the midriff, because nothing else
covers the gap between the breastplate and the leg armour.

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

Armour has a second, larger naming scheme keyed by material rather than race:
`cth` cloth, `lea` leather, `std` studded, `scl` scale, `chn` chain, `plt`
plate, then slot and variant — `pltBody01_04_m.dds`, `pltBoots01_04.dds`. There
are 162 plate textures and 111 chain across the skin archives, so the material
is chosen independently of who wears it.

**DAoC's character skins are not alpha-masked.** The DXT5 faces carry a
smoothly varying alpha channel the engine uses for something else:
`nor_m_head01.dds` has ten opaque texels out of 65536, so reading it as opacity
erases the face. `charconv` flattens it and keeps the colour, which is a correct
skin tone.

## Animation

The character plays DAoC's own recorded animations. A procedural gait remains
as a fallback when they have not been converted, and the HUD says which is
running.

### The `.kfa` files

They are not a separate format. A `.kfa` is a **NetImmerse 4.2.1.0 NIF** — the
same generation as the scenery — containing nothing but `NiKeyframeController`
and `NiKeyframeData`. The parser built for the trees reads them unchanged.

What *is* unusual is how a track names its bone. The nodes in a `.kfa` are all
unnamed. Instead the file's one real node heads two linked lists that run in
lockstep — a chain of `NiStringExtraData` and a chain of
`NiKeyframeController` — and the n-th string belongs to the n-th controller.
That string is a number, which indexes `figures/animnode.dat`, a 178-line list
of Biped bone names:

```
chain position 5  ->  string "6"  ->  animnode.dat line 6  ->  "Bip01 Neck"
```

That indirection is why the animations are portable. A clip names no skeleton
of its own, so any model with Biped bones can play any of them — which is how
one humanoid set serves every player race. The Troll and the Norseman have
byte-identical bone lists, and `animconv -list` scores the same 2081 clips for
both.

Portable does not mean interchangeable, though. A race that moves differently
has its own cycles, and the troll does: `Troll_walk`, `troll_run` and
`Troll_idle` give it a longer, more lumbering stride than `W_VM` does the
Norseman. Clips are converted per character, into that character's own
`anim/` directory, so the model and its movement travel together.

### Finding the right clips

There are 4032 `.kfa` files and nothing maps a model to its animations, so
`animconv -list` scores them instead: it resolves every track through
`animnode.dat` and counts how many land on the character's actual skeleton.

```
scanned 3994 animations, 38 unreadable

animation                          tracks on model  legs duration    keys
W_VM                                   61       61     6    1.20s     685
R_VM                                   61       61     6    0.73s     536
I_VM                                   61       61     6    4.00s    1175
```

2081 clips drive all six gait joints. The naming resolves itself once the
scores are visible: `W_`/`R_`/`I_` are walk, run and idle, and the suffix is a
race-and-sex code — `VM` is Viking Male, which is what `NVikingM.NIF` is. The
scoring is what makes that an observation rather than a guess: clips for other
skeletons, like `I_BritM`, match only 43 of their 93 tracks.

A useful corroboration fell out of it. DAoC walks at 85 u/s and `W_VM` lasts
1.20s, so one cycle covers **102 units** — within 3% of the 105-unit stride the
procedural gait had been using, which was derived independently from the
model's proportions.

### Playing them back

A recorded track supplies a joint's local transform outright, replacing the
bind rotation rather than composing with it — simpler than the procedural path
below, which works in deltas. Rotations are quaternions in NIF's `w,x,y,z`
order, slerped along the shorter arc; a channel with no keys keeps that
component of the bind pose, so a track may rotate a bone without moving it.
Scale needs care: it lives folded into the rotation rows, so rewriting a
rotation silently drops it unless the bind scale is put back.

The clips are authored in place — `Bip01` sways a couple of units but never
accumulates — so the viewer supplies the movement and advances the clip's clock
in proportion to actual speed. That is the same rule the procedural phase uses,
applied to a cycle someone else timed.

### Blending

Changing gait cross-fades over 0.18s rather than cutting, because a foot
forward in one cycle is a foot back in the next and the switch reads as a
twitch.

The fade cannot be done on the matrices `poseClip` writes. Interpolating two
rotation matrices componentwise does not give a rotation: the result shortens
as it goes, so a limb visibly contracts halfway through every transition.
Measured on this model, matrix lerp shrinks a shin by up to **2.5%** at the
midpoint. So a pose here is kept in the same decomposition the file uses — a
quaternion, a translation and a scale per bone, defaulting to the bind value
for any bone the clip does not drive — rotations are slerped along the shorter
arc, and matrices are built once at the end. That holds bone length to
**8.5e-6 units out of 16.3**.

Walk and run hand over their normalised phase when one replaces the other, so
the swing foot stays the swing foot; starting the incoming clip at zero would
cross the legs mid-transition. Idle has no stride to carry, so it starts fresh.
Both clocks keep running during the fade — the outgoing clip finishes its
stride rather than freezing.

Decomposing turned up something about the shipped data. **The bind matrices are
not exactly orthonormal**: the worst bone's rows are 0.9986 long instead of 1,
and two of them are 1.5e-4 off perpendicular. A quaternion can only hold a true
rotation, so the round trip returns a cleaned-up matrix differing from the
original by exactly that much — the test asserts the loss is no larger than the
input's own skew rather than pretending it is zero.

### The procedural gait

Still present as a fallback. A joint's local frame is whatever the artist left it as, so joints are not
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
- **Particle systems are read but not drawn.** Every fixture is placed, but a
  campfire is only its logs and a torch only its bracket: the flame, smoke and
  forge sparks are emitters this pipeline parses and discards. Animated UV
  scrolling (`NiUVController`) and look-at behaviour are read and ignored the
  same way.
- **Two textures are missing from the install.** `BAG.nif` names `mfiga6.dds`
  and `mheada3.dds`, which exist nowhere in the game directory; those three
  props draw white.
- **Three animations per character.** Walk, run and idle are converted and
  cross-fade into one another; the other ~2000 humanoid clips — combat styles,
  emotes, swimming, jumping, death — are reachable by name through `animconv`
  but not wired to anything. Sprint reuses the run cycle, played faster,
  rather than having one of its own.
- **38 of 4032 `.kfa` files do not parse.** They are a small minority and
  `-list` reports them; none is in the humanoid locomotion set.
- **No collision.** The character follows the heightmap and walks through
  anything else, including trees, buildings and the lake surface. The data is
  there to fix it: `nifs.csv` has a Collide flag and a radius per model, and
  the models ship explicit collision hulls that `propconv` currently discards.
- **Two outfits, hand-written.** Any of the 600-odd figure models and 5669 skin
  textures can be combined through `-outfit`, `-shapes` and `-tex`, but the
  pairings are chosen by hand; nothing reads the game's own equipment tables.
  The `fig3/*.mpk` armour meshes are untouched.
- **A troll in plate is not a thing the game would build.** Plate is Albion's
  armour and Troll is a Midgard race, so the live client would never put them
  together. Meshes and textures are independent here, so nothing stops it.
- **No per-race scale.** The troll model is authored 72.9 units tall against
  the Norseman's 71.8 — broader in the shoulders, but the same height. In the
  live game trolls tower over Norsemen, which means the client applies a size
  multiplier this pipeline does not read.
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
