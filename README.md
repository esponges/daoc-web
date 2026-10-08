# daoc-web

A proof of concept: extract Dark Age of Camelot's shipped assets and render
them in a browser. Currently one zone — **100, Vale of Mularn**, its forest and
the village of Mularn — with a playable character you can walk around it: a
**Troll warrior** in Midgard chain with a warhammer and shield by default, or
the **Norseman** via `?char=norseman`.
Wolves and badgers wander the woods and villagers wander the village, each one
built from the game's own creature tables by model number.

Written in Go, with a dependency-free WebGL2 viewer.

## What it does

```
zones/zone100/dat100.mpk ──► terrain.pcx + offset.pcx ──► heights.u16 (256×256 uint16)
                         └─► SECTOR.DAT ──────────────► zone.json    (scale, water, fog)
zones/zone100/lod100.mpk ──► 64 × DXT1 tiles ─────────► atlas.png    (2048×2048)
                         └─► nifs.csv + fixtures.csv ─┐
zones/Nifs/*.npk ───────────► 60 scenery models ──────┴► props.bin   (baked meshes)
                                                      └► props.json  (1006 placements)

gamedata.mpk ───────────────► monsters/monnifs/anims ─┐ (model ID -> figure, skins,
                                                      │  scale, anim set)
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
# With no -anims, animconv converts the clips the game's tables assign the
# race: idle, walk, run, death, combat stance, flinch and three attacks.
go run ./cmd/charconv -outfit troll-warrior   # also converts his hammer and shield
go run ./cmd/animconv -char web/data/char/troll-warrior

# The Norseman is still there; ?char=norseman in the URL switches to him.
# The second run swaps in the locomotion clip-scoring found; index.json
# merges, so his combat clips stay.
go run ./cmd/charconv -outfit norseman
go run ./cmd/animconv -char web/data/char/norseman
go run ./cmd/animconv -char web/data/char/norseman \
  -anims "I_VM,W_VM,R_VM" -as "idle,walk,run"

# NPCs are converted by their model number in monsters.csv.
for m in 56 47 572 153; do go run ./cmd/charconv -model $m; done
for c in large-grey-wolf small-grey-wolf badger norse-male; do
  go run ./cmd/animconv -char web/data/char/$c
done

go run ./cmd/serve                                       # then open localhost:8777
```

Where they stand is `web/spawns/zone100.json`, which is this project's own
placement — see [NPCs](#npcs). `?npcs=0` leaves them out.

Controls: `WASD` to move, drag to orbit the camera, wheel to zoom, shift to
sprint, alt to walk, `C` for the free-fly camera, `F` for wireframe. `Tab`
targets the nearest NPC in front of the camera (again to cycle), `1` switches
auto-attack on and off, `Esc` clears the target. The console exposes a `daoc`
handle — `daoc.warp(60, 70)` puts the character on a heightmap cell, and
`daoc.warp(125, 128)` next to the wolves.

Movement uses DAoC's own rates: 85 units/s walking, 191 running, and the gait's
phase advances with ground covered rather than with time, so the feet keep pace
at any speed.

## Tooling

| command | what it does |
| --- | --- |
| `zoneconv` | zone terrain, texture atlas and water |
| `propconv` | zone scenery: models flattened, placed and grouped by texture |
| `charconv` | one character: mesh, skeleton, skin weights, textures; `-model N` from the game tables |
| `animconv` | recorded animations; by default the model's own anim set, `-list` scores every clip |
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
node web/skeleton.test.mjs troll-warrior  # and for any other converted character
node web/npc.test.mjs                   # ten simulated minutes of wandering
node web/combat.test.mjs                # targeting, fights, deaths and respawns
```

`skeleton.test.mjs` passes for all six converted characters — two playable
outfits, two wolves, the badger and the villager. `npc.test.mjs` loads the real
spawn file and characters against a stubbed-out WebGL context, lays a lake
across the wolves' home, and checks that every NPC wanders, none outpaces its
walk, strays from its radius or gets its feet wet, and all of them pose.

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

`charconv -outfit` names a pairing. Three ship:

| outfit | model | shapes |
| --- | --- | --- |
| `norseman` | `NVikingM.NIF` | `HeadA1,Body1,LBody1,Arms1,Legs1,Gloves1,Boots1` |
| `troll-plate` | `NTrollM.NIF` | `HeadA1,Body5,Lbody4,Arms6,Legs4,Gloves3,Boots5` |
| `troll-warrior` | `NTrollM.NIF` | `HeadA1,Body4,LBody3,Arms4,Legs3,Gloves2,Boots4`, plus a warhammer and shield |

`-shapes` and `-tex Slot=file.dds` override either half for experiments.

**Chain is Midgard's armour.** `pskins.csv` prefixes its armour skins by realm
— `b_` Albion, `h_` Hibernia, `n_` Midgard — and the heaviest each realm wears
runs Albion to plate, Hibernia to scale and Midgard to chain. A Midgard troll
warrior in chain is how the realm kits one out; the plate outfit stays for
comparison. The chain is the tier below plate in every slot, dressed in the one
`chn*` texture family, `05_10`, that has a texture for all of them. The game's
own item-to-armour mapping has not been found, so that pairing is chosen by
eye, with `charshot`.

### Held items

An item is a small static model in `items/` — `items.csv` lists 3000-odd, and
Midgard's include `n_1h_warhammer01` and seven `M_Shield_*` — and it carries its
own attachment markers: empty nodes named `HELD`, `BELT`, `BACK` and `GROUND`,
one for each way it can be worn. The character carries matching sockets on its
skeleton: `Bip01 R Held` in the right hand, `Bip01 L Shield` on the left
forearm, `Bip01 L Back` and `R Back` across the shoulders. Wielding is lining
the two up — the item's `HELD` marker goes where the socket is.

So `charconv` bakes each item's vertices into the frame of its `HELD` marker
and writes it beside the character, under `equip/<slot>/`, as an ordinary
character with one bone. The viewer sets that bone to the socket's world
transform each frame and draws the item with the character's shader; nothing
else is new. The grip checks out from the numbers: in the bind pose `R Held`
has its x axis along the fingers and its z axis out of the front of the fist,
and the hammer's head lies along its own +z, so the handle runs through the
fist, square to the fingers, head forward.

Shields name three textures — `cloakpattern01`, `symbol_001` and the material,
`Shield_Wood` or `sh_metal01` — and the first two are nowhere in the install.
They are the guild's pattern and emblem, composited over the material at
runtime; `items.csv` flags them with "Strip Textures". Without a guild, the
material is what is left, and that is what the converter takes: the first
named texture that exists.

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

## NPCs

### What the client knows, and what it does not

The server names every creature by a number and nothing more. The client turns
that number into something drawable through a chain of tables in
`gamedata.mpk`, which `internal/gamedata` reads:

```
monsters.csv  model 56 "Large Grey Wolf" ─► figure 212, body skin 97, scale 100
monnifs.csv   figure 212 ─────────────────► "wolf" (figures/Wolf.NIF), anim set 7
anims.csv     anim set 7 ─────────────────► walk 241, run 242, idle 240, death 244 …
animnifs.csv  clip 241 ───────────────────► wlf_walk.kfa, loop
skins.csv     skin 97 ────────────────────► wolf.tga, archive 2 (figures/skins/skin002.mpk, as .dds)
```

So `charconv -model 56` needs nothing else: figure, textures, scale and clips
all come from the tables. The same tables settle two questions the hand-built
outfits had left open. **Race scale** is the Scale column: Troll Male is drawn
at 1.30 and Norse Male at 1.09, so the troll stands about a fifth taller —
nothing in either `.nif` records that. And the troll's anim set names exactly
the `troll_walk`, `troll_run` and `troll_idle` that clip-scoring had picked
by hand, which is a reassuring agreement between two methods that share
nothing.

What the client does **not** have is placement. Where a wolf stands, how far it
roams and what it is called in-game all arrive from the server at runtime, and
nothing in the install records them. `web/spawns/zone100.json` is therefore
written by hand for this project: three large and four small grey wolves in the
woods by the lake, three badgers, and nine villagers between Mularn and
Haggerfell. It is not the live game's population and does not try to be.

### Creatures are Bipeds too

The wolf looked like the risk: a quadruped, a strip mesh rather than a shape
list, and a skeleton nobody had tried. It turned out to be 3ds Max Biped in
four-legged form — `Bip01 L HorseLink` between calf and foot, `Bip01 Tail`
through `Tail4` — so `animnode.dat` already names its bones, and every wolf
clip binds all 36 of its tracks with none dropped. The badger shares the rat's
anim set and binds all 37.

`wlf_walk.kfa` is three identical copies of one 1.2-second stride back to back.
The table's frame counts (10 for "Wolf Trot", 54 for "Wolf Walk", same file)
match neither the stride nor the file, so the clip is played whole, which
loops seamlessly.

### Walking at the clip's own pace

A planted foot does not move against the ground, so relative to the body it
slides backward at exactly the speed the body travels. `Skeleton.groundSpeed`
reads that off each walk clip — the median horizontal speed of a foot while it
sits within half a unit of its lowest point — and an NPC walks at that speed
times its scale, with the clip at 1x. Its feet stay planted by construction.

The tables carry "stride" columns that look as though they should give this,
but they do not agree with the clips: the wolf's walk is listed at 52 and its
feet say 40.7. Runs measure less reliably — a run's foot is barely down — so
NPCs only walk for now.

The same reading checks the player's numbers. The troll's walk measures 86.2
units a second unscaled, against the game's walk rate of 85, so the playback
rate calibrated earlier was right; drawn at 1.30x, the clip now runs that much
slower to match.

### Parsing the rest of figures/

Creatures pushed the NIF parser from 444 to **617 of the 633** figure models:

- Six small extra-data tags — integer, integers, boolean, float, colour and
  text keys — stopped 160 figures outright, since a block with no length cannot
  be skipped unread.
- The composite NPC bodies (`albbody01heada.nif` and siblings) set bit 12 of
  the 10.x UV-set field, which means a tangent and a bitangent per vertex
  follow the normals. Without reading them, every float after is skewed.
- Sixteen figures export their weapon and shield attachment nodes —
  `Bip01 R Shield`, `HELD` — with a translation of three NaNs, `7fc00000` in an
  otherwise sound block. The game places those nodes from whatever is held, so
  the value never mattered to it. The parser accepts NaN there and nowhere
  else, and zeroes it.

The sixteen left over are specialised: textures embedded as `NiPixelData`,
`NiTextureEffect` projections on spell-like models, colour and
texture-transform controllers, and two "coco" models.

## Combat

Basic melee, deliberately the simplest version of DAoC's: select a target,
switch attack on, and you swing whenever your swing timer is up, the target is
within reach and inside your front arc. Each swing rolls once to hit and once
for damage. Standing still, you square up to the target; moving, you can still
swing at anything in front of you.

### The clips come from the tables

Combat animation lives in a second table, `canims.csv`, numbered by the same
anim sets as `anims.csv`: medium, high and low attacks, a flinch, a combat
stance, and the weapon-specific variants of each. `animconv` converts nine
roles per character — idle, walk, run and death from one table, the combat
stance, flinch and three attacks from the other — and records whether each
loops. A creature's row names its own clips (`wlf_alo`, `wlf_hits`,
`wlf_grwl`, `wlf_deth` for the wolf; the badger borrows the rat's). A player
race's attack columns are one-handed weapon swings (`a_h_1s_*`), which is what
the troll warrior uses with his hammer; an empty-handed player-race figure
(monnifs.csv's type 1) takes the hand-to-hand set `A_H_H2H_*` instead, and one
with a shield gets a tenth role, `block` (`b_h_1h_high`). What a character
holds is read from its `char.json`, so the clips follow the equipment. All of
them bind every track on every character.

The warhammer's strike is quick: in the middle tenth of the swing the hand
travels 72 units, about 20 m/s, from wound back behind the shoulder to out in
front. The swing ends exactly where the combat stance begins, so the two blend
without a seam.

### Two things animation needed

**Clips that play once.** Everything until now looped. An attack, a flinch or
a death plays through and stops; a death holds its last frame, and the
skeleton checks now confirm each one ends with the head within ten units of
the ground.

**Layers.** A cross-fade swaps the whole body, which would stop your legs to
throw a punch. An action can instead be laid over the upper body, so the
attack owns the torso, arms and head while the run keeps the pelvis and legs.
Where the upper body starts is not obvious: a humanoid Biped hangs its thighs
off `Bip01 Spine`, not the pelvis, so a mask from Spine down takes the legs
with it. The mask starts at `Spine1` and never includes a thigh's subtree; the
combat test checks an upper-body attack moves the running thigh by exactly
zero.

`web/animator.js` holds both, for the player and every NPC alike: a looping
base that cross-fades between idle, walk, run and the combat stance, and a
one-shot action on top.

### NPCs fight back

Hit an NPC and it fights back: it closes at its run speed — read from its run
clip's feet like the walk, within a sane band of it — takes up its combat
stance and swings on its own timer. Its packmates within 900 units join in.
Large grey wolves are aggressive and attack anything that comes within 400
units; everything else only fights back. Chase one more than 2500 units past
its home and it gives up and walks back, healing when it arrives. Killed, it
plays its death, lies 15 seconds, vanishes, and respawns at home 20 seconds
later.

You have 220 hit points and regenerate after six seconds out of combat. With
the hammer you hit for 15–27 every 3 seconds (bare-handed, 9–17 every 2.2),
and the shield blocks a fifth of the blows that come from in front. Die
and every NPC on you goes home; five seconds later you are back on your feet
where you started.

`node web/combat.test.mjs` plays these fights out against the real spawn file
and characters: killing a small grey wolf while its packmates join, standing in
a large wolf pack until it kills you, and running a badger to its leash.

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
- **The playable outfits are still hand-written.** Creatures come from the
  tables by model number, but the two player outfits pick plate and cloth tiers
  by hand; nothing reads `items.csv` to dress a character from equipment. The
  `fig3/*.mpk` armour meshes are untouched. The Norseman also still plays the
  `W_VM` walk that clip-scoring found, where his anim set names `W_hm`.
- **A troll in plate is not a thing the game would build.** Plate is Albion's
  armour and Troll is a Midgard race, so the live client would never put them
  together. Meshes and textures are independent here, so nothing stops it.
- **NPC placement is invented.** The client ships no spawn data, so where the
  wolves and villagers stand is this project's choice, not the live game's.
- **Combat is auto-attack and a shield block.** No styles, spells, parry or
  evade, levels, experience or loot; the hit points, damage, swing times and
  block chance are this project's numbers, not the game's, and the weapon's
  own stats in the item tables are not read. NPCs never flee, and chase in a
  straight line through trees and buildings, as everything here walks
  through them.
- **Equipment is fixed per outfit.** The warhammer and shield are named in the
  outfit; there is no inventory, no swapping, and nothing worn on the belt or
  back, though the markers and sockets for both are there.
- **No target indicator in the world.** The selected target shows in the HUD
  panel only; nothing is drawn under it.
- **16 of 633 figures do not parse** — see [Parsing the rest of
  figures/](#parsing-the-rest-of-figures).
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
