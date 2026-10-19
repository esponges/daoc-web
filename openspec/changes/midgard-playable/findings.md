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

## 5.2 After the hate refactor, in the browser (2026-10-19)

Two Troll Warriors (level 1, classic kit) on the scratch server, at a level 2 spectral hog south of Galplen.

- **Two clients on one creature:** both targeted a spectral hog and fought side by side. After the first died, the second kept fighting one ("The spectral hog misses you!") until they died too. The chat logs do not show whether it was the same hog. The turn to the next player is covered by `TestCreatureTurnsToTheNextWhenOneDies`.
- **Death:** "You have been killed.", then the character came back at the Galplen bind with full health.
- **Leash:** the browser try was inconclusive: the second character's target was out of reach and no blows were exchanged before it teleported away. The leash is covered by `TestAggressiveWolfAttacksAndLeashes`, which passes unchanged.
- **Nothing else differed** from before the refactor.

**Finding: a level 1 Warrior loses to a level 2 creature.** Each character hit the level 2 hog for 1–2 damage (training axe: 1.2 DPS, 2.5 s, 90% quality) and took 10–11 per hit from it, with 39 hit points. Both died; the hog was barely scratched. Task 5.3 checked it.

## 5.3 Low-level melee (2026-10-19)

The fight above is the rule book working as written, and the rule stands:
- **Every number follows the rule book**, and the rule book follows OpenDAoC-Core with its default settings. Mob damage factors 2.5 and 120, a damage factor of 1, PvE melee 1.0, weapon skill 100%, NPC armour `int((1 + L/200) × L × 12.5) / 5` and the player's extra 0.4 AF per level all match, read from its `AttackComponent`, `GamePlayer`, `GameNPC`, `ArmorFactorCalculator`, `WeaponSkillCalculator` and `ServerProperties`.
- **The hog was orange, not yellow.** At level 1 the con band is one level wide (level / 10 + 1), so a level 2 creature is orange to a level 1 character. The first reading of this finding called it yellow, which was wrong.
- **Per hit, at the middle of the variance:**
  - against a yellow level 1 creature (31 HP): the Warrior hits for 4.05 and takes 4.12;
  - against an orange level 2 creature (46 HP): the Warrior hits for 1.93 and takes 10.78.
- **Forty simulated fights** (`TestLowLevelFights`): the new Warrior in the kit beat the yellow creature 20 times in 20 and lost to the orange one 20 times in 20.
- **The source:** a Camelot Herald wiki Combat page, as a search result summarised it (the page itself returns HTTP 402 here), says the ideal low-level melee target is one level below you and a yellow creature can be a tough fight.

On these numbers, losing to an orange creature at level 1 is the classic warning colour doing its job. Two pieces are recorded as rule-book scenarios ("A yellow creature at level 1", "An orange creature at level 1"). Whether 1.65 live was gentler at levels 1–2 than this would need a combat log from the time. None was found, so it stays reference.
