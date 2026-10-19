# Findings

What was seen while applying this change: browser checks, census results and anything that differs from 1.65. No screenshots of game art are kept in git.

## 2.6 The start, in the browser (2026-10-19)

A scratch server (port 8788) ran the new build on a fresh import with templates honoured. The user's live server was not touched.

- A new Troll Warrior appeared in Galplen, Myrkwood Forest, among Krek (Viking trainer), Hrodny (Bonedancer trainer), the weapon merchants Gestod and Helga, and a Troll Flayer guard.
- They stood there for 5.4 minutes with nothing attacking. The chat log had no combat lines.
- All creatures within 4,000 units of the start are level 1 and passive (frogs, hatchlings, snake-finders). The nearest level 26 creatures, Haglion's minions, are about 27,000 units away.
- Teleported 800 units from a level 26 Haglion's minion: its name was drawn in purple, as were the level 14–16 svartalf hunters and mora riders around it.
- The load report said "starts: all 183 imported starts are safe". The start rule replaces nothing on this data.

## 3.2 The kit, end to end (2026-10-19)

A Troll Warrior created on the same scratch server, from a store with `start_equip`:
- wears the training axe (slot 40), the small training shield (41) and bronze stelskodd jerkin, leggings, sleeves, boots, gloves and helm (42–47), and nothing else;
- is bound at the Galplen start.

## 3.3 Levelling gear

The census's levelling-gear table (`docs/parity/README.md`) shows Midgard merchants in regions 100, 101 and 151 selling every kind of Warrior weapon, shields, studded and chain from level 1 to 30, and **nothing above level 30**: 32 gaps across bands 31–50.

The data has 1,302 Midgard weapons and chain pieces above level 30, but only 15 merchant lists sell any of them: 6 placements in region 249 (Darkness Falls), and 2 in region 101. The region-101 ones carry the `OrbsMidgardArmor` and `OrbsMidgardWeapons` lists, which sell only level 51 items for OpenDAoC's own "orbs" currency. That is not classic, and level 51 falls outside the table's bands. A Warrior above level 30 levels with drops, as the data stands. Crafting is a later change. The gap is reported and not filled.
