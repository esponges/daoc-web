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
