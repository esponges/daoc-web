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
