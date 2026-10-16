## 1. Held items

- [x] 1.1 Generate a lookup from the install's `items.csv`: model number to item name and kind (hand weapon, two-hand, ranged, shield), skipping names that cannot be classified; write it to `web/data/` and keep it gitignored
- [x] 1.2 Convert each referenced item model once into a shared folder under `web/data/items/` (reusing the existing item conversion in `cmd/charconv/items.go`)
- [x] 1.3 In the client, attach a shared item to an NPC's `Bip01 R Held` or `Bip01 L Shield` socket when the piece resolves and its kind fits the slot; skip and count otherwise
- [x] 1.4 Test the kind check and the lookup on synthetic rows (no real data); the three-armed-NPC browser check moves to 7.4

## 2. Armour mapping search (one session)

- [x] 2.1 Search `game.dll` and `camelot.exe` for the names of the CSV tables the client reads and see which relate to armour or piece ids
- [x] 2.2 List the install's other archives and tables for anything keyed by the database's armour numbers (47 to 50, 91, 93, 662, 428 to 432 are known test cases)
- [x] 2.3 Compare `figures/` and `items/` filename patterns against known sets
- [x] 2.4 Write each place checked and the outcome into design.md; decide mapping found or not

## 3. Armour drawing

- [x] 3.1 If a mapping was found, generate the armour lookup from it and resolve pieces exactly, using the fallback only for missing numbers (no mapping was found, so this did not apply)
- [x] 3.2 Otherwise implement the tier fallback: tier per slot from the armour type the database gives each worn number, a texture family of that realm, `Color` through `tints.csv`
- [x] 3.3 Leave slots the NPC does not wear to the base model; leave unknown colours untinted
- [x] 3.4 Test the tier choice and the plain fallback on synthetic rows; the browser checks move to 7.4 and 7.5

## 4. Wrap up

- [x] 4.1 Show the skipped-piece count in the developer panel
- [x] 4.2 Update the README: where held items and armour come from, which are exact and which approximate; update CLAUDE.md sources if a new source was used
- [x] 4.3 Reconcile with `npc-equipment`: remove or mark as moved its section 4, then run `go test -short ./...` and the web tests

## 5. Race visuals

- [x] 5.1 Spike: establish from the install how race and gender select a player base model; write where it looked and the outcome into design.md (exact, approximate, or none)
- [x] 5.2 Convert the Midgard player bases, male and female (Troll, Norse, Kobold, Dwarf, Frostalf, Valkyn), by race name from `monsters.csv`; other realms are a later batch and fall back; creatures keep their own model
- [x] 5.3 Draw the race and gender base for players and NPCs that have one; fall back to the current model when unknown or missing
- [x] 5.4 Test the lookup and fallback on synthetic rows (no real data)

## 6. Player gear

- [x] 6.1 Add `race`, `gender` and `eq` to player entities in the protocol (additive), document in `docs/protocol.md`, add a protocol test
- [x] 6.2 Server: add a `gender` column to the character store (existing rows male), load a player's race, gender and equipment and send them in snapshots; send a change when gear changes. No creation UI: the character-creation wizard is out of scope and recorded as a follow-up
- [x] 6.3 Client: draw the local and remote players through the shared item and armour-tier path, updating live
- [x] 6.4 Keep the baked outfit as the fallback; check that with the lookup files absent players draw as today
- [x] 6.5 Update the README and `npc-equipment`'s notes where they say player gear is out of scope

## 7. Validation in the browser (claude-in-chrome)

- [x] 7.1 Start the server and open the client in play; call `tabs_context_mcp` first and use a new tab
- [x] 7.2 Move the character only by dispatching `keydown`/`keyup` from page script with a delay; stay clear of high-level Myrkwood creatures or use the developer panel
- [ ] 7.3 Two characters of different races (and a gender pair) look different; a second client sees the first's race and gear
- [ ] 7.4 Equipping a weapon and armour changes the local player's look live; check three armed NPCs, a full-set NPC, a torso-only NPC and a robed caster in each realm
- [x] 7.5 With the lookup files (and separately `web/data/ui/`) moved aside, NPCs and players draw plain and nothing errors; read the console for errors
- [ ] 7.6 Save screenshots or a GIF in this change folder and note any approximations seen

### Validation notes (2026-10-16, first pass)

Seen in the browser at `/?play=4&zone=105` (a Troll warrior, race 6): the local player loads `data/char/m137` (Troll Male) and is dressed in chain from what they wear; dragging the torso piece out of the inventory shows the base tunic within a frame or two, and putting it back restores the chain (live update, 7.4 for the local player). With `web/data/races.json` and `gear.json` moved aside the player draws as the baked `troll-warrior` outfit with its hammer, NPCs draw plain, and the console shows no new errors (6.4, part of 7.5).

Not yet done: movement by dispatched key events (7.2); two characters of different races and a second client (7.3, only one character exists); armed NPCs, a torso-only NPC and a robed caster (7.4); `web/data/ui/` moved aside (7.5); screenshots are not kept in the repo because they show EA's art (7.6).

Second pass (same day): a temporary female Dwarf test character (race 7, gender 1, set in `players.db`, since deleted) played in a second tab next to the Troll. Her tab loaded the Dwarf Female base (`m193`) for her own figure, and the Troll's tab showed her as a small figure named for her, so race and gender reach the other client (7.3 in part). Movement by dispatched `keydown`/`keyup` turned the camera but barely moved the character in a background tab, so 7.2 is only part done; the figures were too small and partly hidden to judge her armour.

Third pass: the page renders at about 1.3 frames a second in this browser (the offline viewer does too, so it is the environment, not the change), and the game clock steps at most 0.1 s a frame, so movement needs long holds. With 15 s holds of dispatched `KeyW`, `KeyA` and `KeyD` the character turned and walked about a cell each (7.2). With `web/data/ui/` moved aside (and restored) the player is dressed, the plain panels show a player frame and a style tooltip, and the console has only the existing `m405 needs 64 bones` warning (7.5). Full sets, a torso-only NPC (Hord) and cloth-set NPCs were in view; held weapons were too small to confirm by eye at this frame rate, so 7.4 stays open (the lookup resolves Helga, Stein, Gestod, Gudrid and the Troll Flayers to weapon models; Aki's bow and Gord's and Krek's two-handers do not resolve). The account is not an administrator, so `/gm tp` was refused and nothing was teleported. 7.3 stays open for the second client seeing worn gear, and a same-race gender pair.
