# Findings

What was seen while applying this change, and where a source search came up empty.

## 1.1 Attack state (2026-10-21)

The target window now says "(attacking)" beside the target's name, in the stock summary window and the built-in box. Seen on the running server: Browsertroll targeted a mud snake, pressed 1, and the stock window read "mud snake (attacking)". A second press still stops the attack and logs "You stop attacking."; the toggle stays because no source says 1.65 differed.

## 1.2 Attack toggle and auto-walk: no source found

One standard search (1.65 patch notes, Camelot Herald results) found nothing on whether 1.65's attack key toggled or walked the character to the target. Behaviour stays as it is: the key toggles, and attack does not walk.

## 1.3 Night spawns at dusk: no source found

One standard search found nothing on how 1.65 staged night spawns. They keep coming out together at dusk, as the clock brings them.

## 1.4 Pack-to-pack equip

`equip` between two pack slots is the pack's drag-to-rearrange (`web/ui.js` `onMove`), so the server must accept it. Finding 5 of `midgard-playable` came from a script sending the wrong slot, not a fault. `TestPackSlotsRearrange` keeps the behaviour.

## 4.1 Power sources (early search)

A standard search found only that a caster's power pool depends on their primary casting stat (the 1.45c patch notes mention stat items restoring it). No formula or regeneration numbers turned up. Task 4.1 searches further before the rule is written provisional.

## 1.5 Loot (2026-10-21)

A standard search for 1.65 generic loot rates found nothing (only the later 1.104 classic-dungeon loot overhaul). The existing level-banded pool already returns an item at every level from 1 to 40 (200 of 200 draws each), so the thin loot came from the flat 22% chance, not an empty pool. The chance now rises with level (`rules.GenericDropPercent`, provisional). Matching the item to the creature's kind (hides, bones) is not done: the imported items carry no such link. This narrows the proposal's "level and kind"; the spec delta was edited to say so.

## 2.1–2.3 Gender (2026-10-21)

Most of the plumbing was already there: the `gender` column, `fm` on `welcome` and player entities, and female bases for all six Midgard races. The missing piece was the choice at creation. On a scratch server (a copy of the data, port 8788) a female Troll Warrior was made from the login page; the list reads "Female Troll Warrior", and in play the client loaded `m145` (Troll Female) for the player and `m137` (Troll Male) for the Troll Flayer beside them. The night screen is too dark to judge the model by eye, so the proof is which base was fetched.

## 3.1–3.3 Classes (2026-10-21)

- **Races come from the data.** The import's `startup` table lists the races each class has a start for: Berserker 5, 6, 7, 14; Savage 5, 6, 7, 8, 14; Healer 5, 7; Shaman 5, 6, 8; Runemaster 5, 7, 8 (Norseman, Troll, Dwarf, Kobold, Valkyn). The class table uses them.
- **Numbers are provisional.** Primary, secondary and tertiary stats, base hit points and weapon skill base for the five new classes are this project's choice by class kind; a search found only the spec multipliers (Berserker 2.0, Savage 1.5, casters 1.0 on the Camelot Herald). The `character.class-midgard` rule says so. Replace them when a source with the table is found.
- **Everything else was already class-generic**: specs, styles, trainers (a Berserker trains at a Viking trainer, a Healer at a Seer trainer), starting kit and start town are read by class number from the data. Every new class wears its whole kit at level 1.
- **Creation page.** Characters are now offered for every legal race and class pair: `GET /api/me` lists them with ids such as `dwarf-healer`; `troll-warrior` and `norseman` keep theirs.
- **Browser (scratch server, 8788):** created a Troll Berserker, a female Kobold Savage, a female Dwarf Healer and a Norseman Runemaster from the page's API; a Kobold Healer was refused (400). The Berserker entered the world, was levelled to 10 with the admin command and targeted a frog. No blows were exchanged because the target was out of reach, so 3.4 (a fight and styles seen in the browser) stays open and is done in the integration playtest. The stock styles window drew its section titles but no rows on this screen; it is the same window the Warrior uses.

## 4.1–4.2 Power (2026-10-21)

Two searches (the 1.65 power pool and regeneration; the class tables) found only that the pool follows the class's casting stat and the Herald's spec multipliers. The pool is `10 + level × (5 + acuity / 50)` and regeneration a fiftieth of the pool every three seconds out of combat (a two-hundredth in it); both provisional. Spells draw a share of the pool, using the data's `Power` column as a percentage (to be checked against the data when spells are imported, 4.3).

## 4.3–4.7 Casting in the browser (2026-10-21)

On the scratch server with a fresh import (the staged world store now carries the spell lines and each spell's power):
- **Healer** (female Dwarf, level 1): the stock SPELLS window lists three lines (Augmentation, Mending, Pacification) with each spell's level, power and cast time. Clicking Minor Mending logs "You begin casting Minor Mending." and, 2.2 seconds later, "Your Minor Mending heals you for 6 hit points." A shift-click put it on the spell bar, Shift+1 cast it again, and the slot survived a reload. Without `web/data/ui/` the plain panel (P) shows the same book.
- **Runemaster** (level 1): with a sveawolf cub targeted from afar, "Your target is too far away."; after moving close, "Your Lesser Sigil of Ruin hits the sveawolf cub for 6 damage!" and the cubs came at the caster.
- **Not seen:** buffs and a group heal in the browser (they need a second client; the server tests cover them).

Things this showed:
1. **The healed amount is the spell's value, not what was restored.** At full health the log still says "heals you for 6 hit points". Not fixed.
2. **A level 1 caster has a pool of 16 and every level 1 spell costs 1 to 2 percent of it**, so the pool is never the limit at level 1; whether it is later depends on the provisional pool formula.
3. **Spell damage does not scale with the line's spec level** (the data's base value is used); the bolt did 6 damage at spec level 1.
4. **Line of sight is not checked** and the spec says so.

## 3.4, 5.1, 5.2 What the browser reached (2026-10-21)

- **Berserker:** Krek's window offered Axe, Hammer, Sword, Parry and Left Axe with 94 points at level 10; three clicks took Axe to 4 and two style icons appeared on the bar. The Berserker's ability list shows a Protect button, which comes from the data's grants (a Warrior ability); not checked against 1.65.
- **Vault:** Idonna's window opened beside the stock pack.
- **Death:** a level 10 Berserker standing in the svartalf camp was killed in seconds by level 14 hunters hitting for 74 to 79, lost 1 constitution and woke at the bind point.
- **Fights are done in simulation.** The shared screen runs at one to four frames a second, so reaching a creature on foot takes minutes; `TestMeleeClassesFightWithStyles` and `TestRealCasterCreatureHurts` fight on the real data instead.

## 6.1 A group in the browser (2026-10-21)

Two tabs on the scratch server: a Runemaster and a Healer, both level 1, in East Svealand.
- **Group:** `/invite Healina` brought the stock invitation on the Healer's screen; accepting showed both in the stock group window, the leader starred, with class and level.
- **Combat:** the Runemaster teleported to three sveawolf cubs and cast Lesser Sigil of Ruin (6 damage); the cubs killed the level 1 caster before the Healer could heal them, and the group window showed "dead" beside their name. That is the level 1 melee-against-a-caster danger the rules give (cubs hit for 8 to 9 against a Runemaster with 30-odd hit points); a real group would pull one at a time, and 1.65's casters had the same fragility.
- **Not seen:** a heal on another player, a group buff and the Shaman in the browser. The server tests cover the heal on a group member's target and the group buff (`TestGroupSpellReachesMembers`); a heal on a targeted member goes through the same `castTarget` as a heal on oneself, seen above.

What felt unlike 1.65 for the next change (the remaining classes and a real playtest):
1. **Casters die to the first pack.** No pets, no Bladeturn, no armour for cloth casters yet.
2. **Spell damage and heals do not scale with the spec** (data base values only).
3. **The healed amount reported is the spell's, not what was restored.**
4. **No line of sight, no concentration for buffs, no resurrection** (Reception from Valhalla is marked not yet).
5. **Spells that need a corpse, a pet or a group of kinds the slice lacks** show greyed in the book.

## 6.2 Checks (2026-10-21)

`go test ./...`, `go vet`, every `web/*.test.mjs`, the parity census (`docs/parity`: class spell types 13 of 73, creature kinds 16 of 110, 18 provisional rules) and `openspec validate --strict` pass. `TestRideCrossesAZoneBorder` failed once when every package ran in parallel on a loaded machine and passed in all three runs alone and in the package; it is timing-sensitive and was not changed.
