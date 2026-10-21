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
