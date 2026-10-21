# What was seen

Records of what was watched in the browser, with the 2026-10-20 session as the date for everything below unless stated. The data and numbers are from a live server on the final build of sections 1 to 15.

## 16.1 Compass

- **North is the viewer's -y, and that is the zone data's north.** The zones of a region share one coordinate frame: the zone table gives each zone's offset in 8192-unit blocks, and the server's border crossing already uses it (`borderCross`). In Midgard's region 100, West Svealand (offset 80, 90) lies west of East Svealand (88, 90), so +x is east; y grows towards Malmohus (84, 118) and shrinks towards Odin's Gate (72, 70), the usual map with north at the top, so north is -y. The camera's heading is `atan2(dx, -dz)`, 0 when facing -y, and the compass shows heading 0 as N.
- **Facing north.** In Myrkwood Forest (offset 92, 106) a character turned until the compass centred on N and walked on. Gotar (88, 98) is the zone above it, and its x range overlaps Myrkwood's. At y 205 the character was still in Myrkwood; the next check was in Gotar (zone 103).
- **Across the border.** The compass read N, with NW at its left, before the border and N with NW at its left after it, with no turning between. The border crossing reconnected to zone 103.
- **No fix was needed.** The one north held.
- **What this does not prove.** The zone files say nothing of compass points themselves; "east" is the converter's reading of the install's sky file (`east` rings on +x), and the map-style handedness is the usual one. If a later look at the game's own map disagrees, this is the place to look.

## 16.2 Clock

- The stock clock read 6:45 PM while the game time computed from the real clock (real seconds times 24 modulo a day) was 18:45. It advanced as the session ran (7:04 PM, 7:24 PM, 7:47 PM, 12:22 AM later), and the viewer's night look followed it.

## 16.3 Loot and death

- **Loot, a real kill.** A mud snake (level 1) killed in Vale of Mularn opened the stock loot window with 12c and one item row, `Tattered Rawhide Mjuklaedar Sleeves` (`AF 4, level 1`). Clicking the item row took the item (the row went, the coin stayed); TAKE ALL then took the coin. The purse went from 74 to 86 and the sleeves were in the pack. Level 1 creatures usually drop coin only: the imported data has item drops by name for named creatures, and the server adds a 22% chance of a random item of the creature's level.
- **Death, early release.** The character was killed by a pack of young grendelorms. The server set `dead`; `release` sent about a second later stood the character up at the bind point (50823, 30022) with their hit points restored, well before the five second timer.
- **Death, waiting.** With no release the character stood up at the bind point exactly 5.0 seconds after death, with 260 hit points.
- **Release window.** Shown in the page for a character held dead: the stock frame, a timer bar, the text ("Releasing..." once the timer ran out), `/RELEASE` and `/RELEASE BIND` live, `/RELEASE CITY` and `/RELEASE HOUSE` greyed.
- **Limits.** The test browser drew only about 1.7 frames a second here, so the window was seen with the character held dead on the client; the server side (early release, five second wait) was timed from the page's own state, not from a screenshot.

## 16.4 Icon numbers

One bounded search, three web searches for a table mapping the community database's icon numbers to this install's icon sheets (including `Two Fists` 105 against sheet 636, and `Bloodletter` 60 against 505). They returned only 1.65 patch notes and general icon-extraction tools; no table, tool or write-up of the client's numbering. None was found, so styles and items stay drawn by weapon line and kind, as in `archive/2026-10-19-stock-ui-followups/design.md`.

## 17.3 Interior fit

- **What changed.** The fit is made against the placed models' vertices, so the middle of a large floor or wall counted as off the geometry. A zone that comes out under nine in ten of its points on the geometry or inside a placed piece's oriented box (a light in the middle of a cavern) is now fitted again against the triangles sampled across, and the better fit is kept; a zone that fits well keeps its first fit (sampling everywhere moved Spindelhalla's origin 82 units across the flat top of its fit and broke the spawn-floor check, so it is not used for them). Nisse's Lair went from 77 to 80 of 86 (93%); Trollheim from 87 to 90 on the geometry and 94 of 138 with the boxes; Tuscaran Glacier's first fit stood (33, 41 with the boxes).
- **What did not change.** Of Trollheim's 48 points off the geometry, 4 are within 128 units of it, 20 within 512 and 34 within 1024; they spread across the whole zone (x 13456 to 33512, y 3910 to 42957). Fitting the misses alone under each of the eight orientations (four turns, mirrored or not) over every origin in the zone put at most 27 of the 48 on any geometry (the same orientation, an origin 2100 units west): no second frame. Tuscaran Glacier's 17 misses behave the same (10 at best). So parts of those two dungeons are missing from their layouts, and no transform recovers them.
- **The test** (`TestTransformsFit`) holds every Midgard interior at nine in ten on the geometry or inside a piece, and holds Trollheim at 0.65 and Tuscaran Glacier at 0.78, where they land.

## 17.1 NPC faces

- **Cause.** The converter read the `Head` column of `monsters.csv` as the head's skin. For a humanoid that column is the hair colour (`hair_white.tga`, `hair_black.tga`): `Frostalf Male 1` has Head 1016 `hair_white`, and a separate Face 1657 `SV_MHed1`. The head shape was drawn in a flat hair texture. `gamedata` now reads the Face column (column 18) and `charconv -model` dresses the head in it when the row has one (a monster with only a Head skin keeps it). All 133 converted humanoids and races were converted again; no failure.
- **Seen.** With `charshot` on the converted dwarf guard (model 186), the smith Genlu Edrill (234) and a huldu hunter (127), at three times zoom on the head: each has a face, with eyes, nose and mouth in a skin tone, and the huldu hunter's hair is on its head; before, the same shapes were a flat colour. In the live page in Vale of Mularn the teleporter Stor Gothi Annark shows a face and fair hair, and Yver Tiu a bald, skin-toned head. The guards themselves stood out of the camera's sight behind buildings, so the guard and the smith were looked at with `charshot`, not in the page.
- **Plain NPCs.** With `web/data/gear.json` and `races.json` moved aside, the page loaded and every NPC and both characters still drew; they came out as their plain base models and the baked outfit. Both files were put back.
- **Limits.** Hair is not drawn as its own piece (the `Hair` column is a style number, and the shapes that carry it were not mapped), so a head with hair in the Head skin shows it painted or not at all.

## 17.2 Race and gear (partly done)

- **Two races.** `Browsertroll` (Troll Warrior, race 6) and `Secondtroll` (Norse Warrior, race 5) stood next to each other in Vale of Mularn, each client in its own tab. Each saw the other's race by body: the troll tall in chain, the Norseman shorter with a bare head, the labels `Browsertroll 12` and the other's name. It was night in the game, so the picture is dark.
- **Equipping live.** From the page, the Troll's worn arm piece was swapped for `Tattered Rawhide Mjuklaedar Sleeves` and back. The server's `eq` for that slot went from model 232 to 282 and back, in the Troll's own client and, within about a second and a half, in the Norseman's client, which sees the entity's `eq`. The difference in the dark picture is too small to point at, so what was checked is the data both clients drew from.
- **Not done.** A same-race gender pair: nothing lets a player choose gender yet (characters are male), so no female character exists to compare. "In each realm": only the Midgard races and gear are converted, so Albion and Hibernia NPCs cannot be looked at, and the three armed NPCs, full-set NPC, torso-only NPC and robed caster were not each looked at. These need the character-creation wizard and the other realms' conversions.

## 17.4 Interior tour

All 13 Midgard interiors were opened in the offline viewer (`/?zone=N`), each looked at three ways: where the character stands at the start, a walk forward with W held for four seconds, and a free-fly plan view from above (`C`, then the camera over the zone). The test browser draws only a few frames a second, so a walk covered 57 to 76 units in four seconds; what the walk shows is that the feet stay on a floor.

| Zone | Name | Start | Walk (units, change in height) | Plan view |
| ---: | --- | --- | --- | --- |
| 120 | Jordheim | in a street between walls, lit | 57, none to speak of | roofs and streets, lit by torches |
| 125 | Spindelhalla | on a wooden platform at the entrance, dim | to 23499, 23558 | cave pieces, warm lights |
| 126 | Vendo Caverns | on a ledge in the dark | 60 | cave chambers joined by tunnels |
| 127 | Varulvhamn | in a rock passage with a lit end | 57, -25 | chambers and passages |
| 128 | Cursed Tomb | in a stone corridor | 76, +39 | round and square rooms |
| 129 | Nisse's Lair | on a rock ledge, dark | 57, -9 | chambers (this one is now the fit of 80 of 86) |
| 149 | Nyttheim | on dark ground among plants | 57, -25 | large round shapes, very dark |
| 150 | Trollheim | in a cave mouth, dark | 76, none | a large round chamber and a passage, seen from the fit's centre |
| 160 | Tuscaran Glacier | nearly black: the character is barely seen | 76, none | blue-grey ice chambers, seen from the fit's centre |
| 161 | Iarnvidiur's Lair | in a rock passage, red lava glow | 57, -13 | rock with red light |
| 188 | Midgard's Darkspire | almost black | 60, -3 | almost black |
| 189 | Midgard's Glashtin Forges | in a hall, dim | 60, -2 | a ring of platforms around a centre |
| 226 | Midgard's Abandoned Mines | on a lift platform with ropes and barrels, dim | 57, none | only torches and a few props show |

- **Frame rate.** The viewer's own counter read 21 to 23 frames a second in every interior, 22 in Abandoned Mines (978 placements, about 29,000 triangles in the window's figure). The test browser's page loop ran at 1 to 4 frames a second, so the counter is the better figure and neither is a measure of a player's machine.
- **Findings.**
  1. **Tuscaran Glacier and Darkspire are nearly unlit at the start.** The spawn is far from any light (Darkspire has three lights in all, Tuscaran 49 for a very large zone). Nothing is wrong with the geometry; the interior is as dark as its light data. The ambient is 0.27; a first-time player there sees almost nothing.
  2. **Trollheim and Tuscaran Glacier have far-off pieces.** Their geometry bounds run tens of thousands of units outside the region the exits and lights lie in (Trollheim from x 9453 to 62345), so a plan view from the bounds' centre sees nothing, and the missing lights of 17.3 are not explained by them. Looked at from the fit's centre both draw their chambers.
  3. **Spindelhalla's start is a narrow platform** at the entrance with a wall close in front; the chase camera is pulled in tight. It stands on a floor.
  4. **The plan view is of limited use** in the large, sparse zones (Abandoned Mines, Darkspire): the geometry is a thin shell and the plan view shows mostly black. A walk along a passage is the better look.
- No zone's start was off a floor, none fell during a walk, and none failed to load.
