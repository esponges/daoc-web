## Context

From the `npc-equipment` spike (see its design.md):

- Armour is numbered per set (a guardsman: legs 47, arms 48, hands 49, feet 50, torso 662, helm 93, cloak 91). Only some numbers appear in `pskins.csv`. The same numbers exist in `skins.csv` and `mskins.csv` but name unrelated rows (`mskins` 47 is a chain body texture, `skins` 47 a Saracen head), so number equality across tables is coincidence.
- `items.csv` (ID, name, NIF) resolves many weapon numbers sensibly (311, 315, 444) and some wrongly (79 is a bow in a hand slot; 3, 4, 19, 22 are absent).
- `Color` is an id into `tints.csv` (189 rows of RGB).
- Today held items are baked into a character's own folder (`web/data/char/<name>/equip/<slot>/`) by `charconv` and attached to sockets in `web/character.js`. That suits one player outfit, not thousands of NPCs.

## Goals / Non-Goals

**Goals:** weapons and shields on NPCs that fit their slot; a bounded, recorded search for the real armour mapping; a tier fallback that looks plausible and is labelled approximate; no data committed.

**Non-Goals:** weapon effects, emblems, heads and faces, and guessing a numeric offset to force a match (forbidden by CLAUDE.md).

## Decisions

1. **Order: weapons, then search, then fallback.** Weapons and the client plumbing are the same whichever armour source wins, so they ship first and are not held up by research that may fail.
2. **Shared item models.** `charconv` (or a new subcommand) converts an item model once into `web/data/items/<name>/` and the client attaches it to the NPC's socket at runtime. The existing baked `equip/` path stays for player outfits.
3. **Kind check by name and slot.** Each `items.csv` NIF name carries its kind (`_2h_`, `Shield`, `_bo_`, `_st_`). The lookup generator classifies the name into hand weapon, two-hand, ranged or shield and stores it, and the client rejects a piece whose kind does not fit the slot. Names that cannot be classified are skipped.
4. **Bounded armour search.** One working session, in this order: (a) search `game.dll` and `camelot.exe` for the names of the CSV tables to learn which the client reads armour from; (b) look in the other archives under the install for tables that mention armour or piece ids; (c) compare the `figures/` and `items/` filename patterns against known sets. Record each place checked and the result in this file. If a table is found, generate the lookup from it. If nothing is found by the end of the session, stop and use the fallback.
5. **Fallback armour.** Revised after the search (see below): the armour type of each worn number comes from the imported item table, not from the NPC's realm. `charconv -gear` writes `armor` into `gear.json`: per `slot:model` a level 0 to 4 (cloth, leather, studded or reinforced, chain or scale, plate) and a realm, per `realm letter:level` a texture set chosen from `pskins.csv` and extracted to `web/data/armor/`, and the RGB of every colour in use from `tints.csv`. The converter bakes every tier of each slot into a humanoid creature (shapes tagged `slot`, `tier`, `base`, bounds unchanged) and the client shows the tier for the level, in that set, tinted. Slots the NPC does not wear keep the base model. The fallback is labelled approximate in the README.
6. **Where the lookup lives.** A generated JSON in `web/data/` (gitignored), built from the install, mapping database model number to `{kind, item}` for held items and, if found, to armour shapes and textures. Absent file means NPCs draw plain.

7. **Players share the NPC path.** The server sends `race`, `gender` and `eq` for players on the same terms as NPCs (additive fields, documented in `docs/protocol.md`, with a protocol test). The client resolves player gear with the same held-item and armour-tier code, so there is one resolution path. The fixed baked outfit (`charconv -outfit`) stays only as the fallback when race, gender or a piece cannot be resolved. Equipping or unequipping updates the look live.
8. **Race first, by spike.** Before converting bases, establish from the install how race and gender select a base model (character tables, `figures/` patterns, the `fig3*` tables seen in `game.dll`). If no exact mapping exists, record that, use the nearest model the install gives, and label it approximate. No numeric offsets (CLAUDE.md).
9. **Validate in the browser.** After implementation, claude-in-chrome drives the real client (server running, character in play). Keys are sent as `keydown`/`keyup` dispatched from page script with a delay, and the check stays clear of the Myrkwood spawn's high-level creatures (or uses the developer panel, backquote) to avoid dying. Evidence is screenshots or a GIF kept in this change.

## Risks / Trade-offs

- Player race ids in the database may not match the install's models; the race spike may only partly succeed, as the armour search did.
- Converting a base per race and gender is long; convert the player races first and fall back to the current outfit for the rest.
- The search may find nothing, leaving armour approximate. Accepted, and labelled.
- Some weapons will be wrongly resolved where the kind happens to fit but the model differs (a mace for a different mace). Acceptable; they look plausible and are visible only up close.
- Tinting a texture family by multiplication can look muddy for dark colours. Check against a few NPCs in the browser and clamp if needed.
- Shared item models need the NPC's socket names to exist on every creature model. Creatures without them are skipped.

## Open Questions

- Does the client read armour from a table in another archive, or from code?
- Which texture families exist per realm for every slot and tier? The README notes only one family, `chn*` `05_10`, covers all Midgard chain slots.
- How do race and gender select a player base model in the install?
- Do any NPC creature models lack the `Bip01 R Held` and `Bip01 L Shield` sockets?

## Findings while applying

- On a scratch server over the real import, the lookup resolves 72 of 466 held-item numbers (187 absent from `items.csv`, 207 whose kind does not fit the slot). Many rejected numbers land on props, helms and map markers (`int_floor_bed_01`, `nhelm2`, `pos_bluemarker_area`), which confirms the database's numbers and the install's rows are not aligned for much of the range.
- Resolved weapons can still be wrong. Midgard Viking Dreng NPCs resolve to Hibernian items (`h_2h_direclub01`, `h_2h_elvengreatsword01`), and a Dwarven Guard to an elven great sword. The kind check cannot catch a plausible weapon of the wrong realm. Candidate follow-up, not done: reject a resolved item whose realm letter (`b_`, `h_`, `n_`) contradicts the NPC's realm. A numeric offset is not an option (CLAUDE.md).
- Guild text and per-slot armour pieces arrive and the guild label renders in the browser. Held items resolve per NPC and were seen drawn: a rattling skeleton in open ground resolved to `h_2h_elvengreatsword01` on `Bip01 R Held` and the browser showed a long blade in its hand, while unarmed skeletons beside it showed none. That is one armed NPC seen, not the three task 1.4 asks for. Armed NPCs in the villages and forts stayed out of reach of the follow camera (walls, roofs).

## Armour mapping search: outcome

Checked, in the order of design decision 4:

1. **Client binary.** `game.dll` names every table the client loads. Armour-related ones are `pskins`, `mskins`, `skins`, the `fig3*` family (`fig3map`, `fig3parts`, `figtofig3main`, `fig3tints`), `meshmateriallist`, `nifs`, `textures`, `nifproxy` and `texproxy`. `camelot.exe` and `camelot.bin` name none. `nifs.csv`, `textures.csv`, `nifproxy.csv` and `texproxy.csv` are in no `.mpk` under the install and in none of `mft.myp`, `patch.myp`, `main.dat` or `camelot.bin` (string search), so the table that resolves an armour model number is not available to us.
2. **Direct id matches, all negative.** The database's armour numbers are not `pskins` ids (688 to 697 are Celtic scale and leather textures there, plate in the database), not `skins`/`mskins` ids (the same numbers hit unrelated rows), and not `fig3parts` ids (47 is a hair piece).
3. **Filenames.** `figures/` has 637 NIFs and `items/` 3392; humanoid NIFs carry their armour tiers as shapes (`Body1` to `Body5`, `Boots1` to `Boots5`, and so on). No filename encodes the database's numbers.

**Result: no mapping found, so armour is approximate.**

**What the database does give exactly.** Armour numbers are item model numbers, and the imported `item` table says what each belongs to. Model 47, 48, 49 and 50 are the legs, arms, gauntlets and boots of Albion plate (object type 36, realm 1); 428 is a Hibernian scale hauberk (38, realm 3); 245 a Midgard cloth vest (32, realm 2). Sets are blocks of five consecutive numbers (torso, legs, arms, hands, feet). Of 879 (slot, model) pairs, 475 belong to one armour type and realm and 404 to several. So the fallback can read the armour type of a worn piece from the database instead of guessing it from the NPC's realm: take the most common armour type among items using that number in that slot, and use the realm only to choose the texture family.

**Cost found while reading the client side.** A converted creature such as m286 (`HFirbolgM.NIF`) contains only the tier-1 meshes of each slot (`Body1`, `Legs1`, ..., textured `f_cth0_*`). Drawing chain or plate needs the higher-tier meshes baked into the creature's mesh and chosen per NPC, plus a per-instance texture and tint. That is converter and renderer work, not just a lookup.

## Race and gender mapping spike (5.1): outcome

Checked `gamedata.mpk`'s `monsters.csv`. Player races are ordinary rows there, named `<Race> Male` and `<Race> Female` (plain `Briton Male`, numbered variants `Troll Male 2` to `8` for heads, `Merchant` and `Shopkeeper` variants for NPCs). Each row gives the figure (`#` column: Briton 206/209, Highlander 210/211, Saracen 213/214, Avalonian 217/218, Troll 237/238, Norse 239/240, Kobold 241/242, Dwarf 243/244, Celt 250/251, Elf 252/253, Firbolg 254/255, Lurikeen 256/257, Sylvan 346/347, Inconnu 348/349, Valkyn 351/352, Half Ogre 449/450, Frostalf 464/465, Shar 490/491, male/female), the body-slot skins (Body, Head, Arms, Gloves, Lbody, Legs, Boots, Cloak, Helm) and the race scale. That is the same chain `charconv -model N` already uses for creatures.

**Result: exact, by name.** The install names the race and gender of every player model, so a (race, gender) pair selects a row by matching the race's name, never by a numeric offset. The one thing that must be written down by hand is the database's race number to race name (a fixed list from the game's own races, kept in code beside the archetypes). What the server lacks is gender: `store.Character` has race and class only, so a gender column and a choice at creation are part of 6.1 and 6.2. The existing characters have no gender and are treated as male.

Open for 5.2: which head variant to use (the numbered rows differ only in head), and whether the bases below the two warriors already converted (Troll and Norseman) share the skeleton and clips of their race.
