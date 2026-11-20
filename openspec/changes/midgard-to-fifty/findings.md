# Findings

Part A sources, read 2026-11-19 from DOLSharp (`Dawn-of-Light/DOLSharp`) and OpenDAoC-Core through the GitHub API, in the session scratchpad and not in this repo. Formulas are restated in our own words; no code or identifiers are carried over. File names are for finding them again (under `GameServer/`). Part B to D findings are added when those parts begin.

## 1.1 Part A

### Line of sight — `spells/SpellHandler.cs` (both projects)

- Both servers check line of sight to the target of a spell when the cast begins and again when it lands, and refuse or fail it with a message. This server's message is "You can't see your target!" (the live client's own text, kept in the spec).
- Neither checks it for a spell on the caster, on the group, or an area spell centred on the caster. An area spell that is centred elsewhere checks each victim.
- What blocks is each project's own geometry (their zone meshes); ours is the hull, trunk and heightmap data we already load.
- The eye height and the storey rule (floors count only between storeys) are ours: the community servers do not publish theirs. Status: provisional for the geometry, reference for when it is checked.

### Concentration — `propertycalc/MaxConcentrationCalculator.cs`, `ECS-Components/EffectListComponent.cs`, `gameutils/EffectHelper.cs`, `serverproperty/ServerProperties.cs`

- **Pool.** DOLSharp and OpenDAoC agree: a base of 4 × level × 2.2, adjusted by the caster's casting stat (acuity, piety, intelligence or charisma, by class), then halved. The stat factor is a percent: a stat over 50 adds (stat − 50) / 2, one under 50 takes away (50 − stat). A class with no casting stat has no limit; the Warrior and the other melee classes have no spells, so no pool in practice. At level 50 and stat 100 that is 275, at stat 200 it is 385.
- **Not by spec.** Mythic's Grab Bag answer says concentration does not depend on the spec, but on level and the casting stat (Camelot Herald, Grab Bag Q&A).
- **Refusal text.** "This spell requires N concentration points to cast!" is the community servers' message and is used here in place of the first draft's.
- **Range.** OpenDAoC disables a concentration buff while the target is farther than 5,000 units from the caster, and turns it on again on return; the endurance-regeneration buff has 1,500. DOLSharp has no range rule. This server takes 5,000 units, and ends the buff when the target stays out of range for a grace time of 20 seconds; the grace time is ours so that a short run past the edge does not cost the buffs. Status: reference for the range, provisional for the grace time.
- **Limit of 50.** DOLSharp also caps a caster at 50 concentration spells at once. Not needed here: the pool is the limit.
- **Death and zoning.** Both end a caster's concentration effects when the caster dies. Status: reference.

### Resurrection — `spells/ResurrectSpellHandler.cs`, `gameobjects/GamePlayer.cs` (both projects), and the data's own spell text

- **The offer.** The target is asked "Do you allow X to resurrect you with N percent hits?" and has 15 seconds to answer; no answer is a decline. A decline gives the caster their power back. The Camelot Herald's patch notes (1.37c, December 2001) give the same 15 seconds. Status: documented.
- **What is restored.** The target stands up where the caster is, with `ResurrectHealth` percent of their hit points and `ResurrectMana` percent of their power. The spec lines' resurrections restore endurance by the same percent; a base-line one restores none. Status: reference.
- **Experience: the first draft was wrong.** Neither community server gives experience back. The data's own spell text, which is Mythic's, says a resurrection "eliminates the experience penalty and con loss they would have suffered were they to have /release". In our two-stage death (the first loss at death, the same again on release, `rules-death`) that means a character who is raised skips the release stage: no second loss, no constitution loss and no gravestone. What was lost at death stays lost. The spec is rewritten to say so. Status: documented (the spell's own text) for the shape; the first stage is the existing rule.
- **Sickness.** A raised character carries resurrection sickness: their effectiveness (the damage they do and the strength of their spells) is lowered. The data gives the spell (id 2435, "Resurrection Illness") a duration of 60 seconds and a value of 30 (percent); the RvR sickness is 60 seconds at 50, and RvR does not apply here. DOLSharp's code halves the sickness's duration after a full (100 percent) resurrection, with a comment citing patch 1.56; patch 1.52 gave back the concentration a character loses while sick. Status: reference for the 30 percent and the 60 seconds (the data), reference for the halving (one comment in DOLSharp; no patch note was found).
- **Power cost.** The community servers charge the caster a share of their maximum power that falls as the target's level falls below theirs (half of it for an equal level). The data carries a cost on each resurrection spell (a negative number is a percent of maximum power); we use the data's.
- **Release.** A dead character may release after 10 seconds and is forced to release after 15 minutes (900 seconds); patch 1.65 added the on-screen timer for it. The two numbers are DOLSharp's. Status: reference.
- **Realm points** come back to the resurrector in the community servers; there is no RvR here, so it is not used.

### Buff numbers — the imported data (checked 2026-11-19 against the community dump's `spell.sql`)

- **Concentration buffs** (`ConstitutionBuff` 23 rows, `DexterityBuff` 34, `AcuityBuff` 15): target Realm, range 1,000, cast time 3, duration 0 (no timer), power 0, and a `Concentration` cost of 2 to 8. `Value` is the number of stat points (14 for Blessing of Health, 32 for Acuity of the Prophet). The cost is held from the pool, not paid in power.
- **Haste** (`CombatSpeedBuff`, 79 rows): `Value` is read as the percent of melee swing speed gained (16 on the ones sampled), with a duration of 30 to 600 seconds and `Concentration` 0 on the rows sampled (timed buffs). The handler was not read; the meaning of `Value` is read from the spell descriptions in the data. Status: provisional.
- **Damage shield** (`DamageShield`, 97 rows): `Damage` is the damage the shield deals to an attacker (a spell damage value that scales with the caster's spec like other damage), `Value` is 0 on the sampled rows; duration 1,200 seconds. Whether it scales with the caster is not settled by any source read. Status: provisional.
- **Endurance and power regeneration** (`EnduranceRegenBuff` 27 rows, `PowerRegenBuff` 33): group pulsing spells of the bards' songs and chants, `Value` the points returned on each tick (1 or 2), duration 5 to 6 seconds, `Pulse` set. They ride on the existing pulsing songs.
- **Cures** (`CurePoison`, `CureDisease` 3 rows each; `CureMezz` 7; `CureNearsightCustom` 2): `Duration` and `Value` 0; the power cost is negative, a percent of the caster's maximum (−6 is 6 percent). Each removes the lasting effects of its kind from the target (see the debuff kinds that come in part C; until then, the mesmerize cure acts on the existing mesmerize).
- **Spread heal** (`SpreadHeal`, 9 rows): a heal split among group members, weighted to those lowest on health. The rows were not decoded by hand, since their text fields hold commas; the importer reads their `Value` and `Damage`. Status: provisional.

## 1.2 Searches for a published 1.65 account

All run 2026-11-19 as web searches.

| Question | Query | Result |
| --- | --- | --- |
| Resurrection sickness and experience | `Dark Age of Camelot resurrection sickness experience returned resurrect spell percent health power duration 1.65` | No DAoC page with the numbers. Camelot Herald patch notes: 1.37c (accept within 15 seconds), 1.52 (concentration restored after sickness), 1.56 via DOLSharp's comment (a full rez halves the sickness); Allakhazam's Revive entry (10 percent health). |
| Concentration pool | `Dark Age of Camelot concentration points pool formula level range 2000 units buff cancelled concentration disorder.dk` | No page with the formula. Mythic's Grab Bag answer: spec does not matter, the casting stat and level do. |
| Forced release | `DAoC release forced after 10 minutes death release timer ...` | Patch 1.65 added a timer showing the time to automatic release; its length is not given. DOLSharp's 900 seconds is used. |

## 1.3 Settled

1. **Experience on resurrection:** no experience is returned; the release stage (second loss, constitution loss, gravestone) is skipped. Documented by the spell's text.
2. **Sickness:** 60 seconds, effectiveness lowered by the data's 30 percent, halved in time by a full resurrection. Reference.
3. **Offer:** 15 seconds to accept; release stays possible; forced release at 900 seconds, not before 10. Reference.
4. **Pool:** 4.4 × level, adjusted by the casting stat as above. Reference.
5. **Concentration range 5,000 units, grace 20 seconds.** Reference for the range, provisional for the grace time.
6. **Line of sight:** checked at start and landing; not for self, group or caster-centred area spells; provisional geometry.

## 2.4 and 4.5 What was seen in the browser (2026-11-19, scratch server)

A scratch server with the new build and a fresh import (the concentration, resurrection and focus columns present), a level 10 Healer and a level 6 Dwarf Warrior in Myrkwood Forest near an isolated haunt, and a second Healer.

- **Release window:** a Healer killed by a creature saw the stock Release window with "You will be released in 15 minutes" and its bar and buttons; a dead character is no longer stood up after 5 seconds.
- **Concentration window:** the K key opens the list on the stock dialog frame; with no buffs held it reads "Concentration 0 / 49" for the level 10 Healer once the pool has been sent, and "You hold no buffs." Before the first stats message it read 0 / 0: the pool is now sent with the welcome, so the first look is right (fixed; checked by the world test of the stats message, not seen again).
- **Not observed:** casting a concentration buff and cancelling it from the list; the buff ending out of range; a Healer dying with buffs; the resurrection prompt and the raise; casting through a wall, past a tree and over a hill; a dungeon caster coming round a corner. Every attempt was lost to the setup rather than the game: a creature killed the first Healer before it could cast (the character's saved hit points were those of level 1, so it was frail), and with two clients in one Chrome window the background tab runs at about one frame a second, so a second character cannot be walked or aimed (the game has no command to target a fallen player, and Tab only cycles creatures). Covered meanwhile by the world tests of each rule, by the real-zone tests of line of sight (Mularn: 121 of 4,000 lines of 600 units blocked, 2 microseconds each; Spindelhalla: walls and floors block, a ledge is seen over) and by the node tests of the windows.
- **Found:** an offer prompt, a plain release button and the concentration list have no way to be driven from outside the page, so a visual check of the resurrection prompt needs a second real browser window; tasks 2.4 and 4.5 stay open for it.

# Part B: merchant gear above level 30 (read 2026-11-19)

## 5.1 Sources and what the data holds

| Question | Searched | Found |
| --- | --- | --- |
| What quality and level range 1.65 Midgard merchants sold | `Dark Age of Camelot merchant sold weapons armor quality level cap Midgard 1.65` | No source gives a quality or level cap for merchant stock. The patch notes only say vendor weapons were corrected for damage, speed and weight. The project therefore keeps the data's own pattern and marks the table **provisional**. |

What the community data shows (queried in the imported world store, not copied):

1. **Every piece a Midgard merchant sells up to level 30 is quality 85, with no bonuses.** The few higher qualities in merchant lists above level 30 carry bonuses and are not generic gear.
2. **The data has generic Midgard gear to level 50.** Items of level 31 to 50 with no bonus, no class lock, a lower-case name (named drops are capitalised) and a price exist for every kind a Midgard class wears or wields. For armour, weapons and shields at quality 85 they cover every band. Swords, hammers and axes at 36 to 45 exist only at quality 100 (the adamantium and asterite lines); the rule takes the lowest quality that exists in a band, so these bands are sold at 100.
3. **Where the data's lists live:** Jordheim has a merchant for each kind (a weapon merchant for swords, hammers and axes, spear, bow, thrown, claw and axe, staff and shield merchants, cloth, leather, studded and chain merchants). The frontier towns share several of these lists.

## 5.1 Settled

1. **Rule (provisional):** per kind of piece and band of five levels, the data's generic items (Midgard or neutral, level 31 to 50, price over 0, no bonus level, stat or proc bonus, no class lock, lower-case name) at quality 85, or the lowest quality present where it has none.
2. **Where:** each object type is added to one of Jordheim's existing merchant lists (`importer.AddedList`), after the list's last page. The lists Jordheim shares with other towns (shields) carry the stock to them.
3. **Marked:** each added row has `merchant_stock.added` set to the reason; the census prints a star beside a count that only the project's stock fills.
4. **Result:** 520 rows added; the census gap count goes from 32 to 0.

# Part C: debuffs, creature buffs, wards (read 2026-11-19)

Read from OpenDAoC-Core's spell handlers and effects through the GitHub API (`GameServer/spells/*`, `GameServer/ECS-Effects/*`), kept in the session scratchpad and not in this repo. Restated in our own words; no code carried over.

## 6.1 What each spell type does

- **Stat, armour and resist debuffs** (`StrengthDebuff`, `DexterityDebuff`, `StrengthConstitutionDebuff`, `DexterityQuicknessDebuff`, `*ResistDebuff`, `BaseArmorFactor` and armour absorption): a negative buff of `Value` points (a percent for the resist, absorption and speed kinds). Duration is the spell's, lengthened by the caster's spell-duration bonus, shortened by the target's resist to the spell's damage type, never under 1 second and never over 4 times the spell's. Debuffs do not stack with another of the same kind. Status: reference.
- **Combat speed debuff:** the target's attack speed is slowed by `Value` percent (their swing takes longer). Reference.
- **Disease:** the spell's text says it all: slows the target by 15 percent, lowers its strength by 7.5 percent and halves healing it receives (`inhibits healing by 50%`). Duration is shortened by the target's resist. Reference (numbers from the spell text).
- **Nearsight:** lowers the target's spell and archery range by `Value` percent. A target that already has it, or has just had it end, is immune for a time; the spell also interrupts the target. Reference; the immunity time is provisional (no number found).
- **Confusion:** against a creature, `Value` percent chance on landing to take hold (a negative value is a 100 percent chance, with `|Value|` percent to attack an ally); every 6 seconds the creature drops its target and picks a random enemy within 750 units. Against a player it interrupts the cast or shot in progress and does nothing more. Reference.
- **Amnesia:** against a creature, the creature forgets who it was attacking, and any spell a player is casting at it is negated. The chance is the spell's own column. Reference.
- **Bladeturn:** a bubble that absorbs the next melee hit, then ends. Reference.
- **Ablative armour:** a health buffer that takes `Damage` percent (25 if absent, capped at 100) of each physical hit until `Value` damage total is spent. The magic and both-kinds variants do the same for spell or all damage. Reference.
- **Life transfer:** the heal is twice `Value` percent of the caster's maximum hit points; the caster pays half of it, and the heal is halved on a diseased target. The caster is never taken below 1 hit point. Reference.
- **Damage with debuff** (`DirectDamageWithDebuff`, `DamageSpeedDecrease`): the damage of the spell plus the debuff or slow of `Value` percent. Reference.
- **Unbreakable slow** (`UnbreakableSpeedDecrease`): a speed decrease that damage does not remove. Reference.
- **Fatigue consumption debuff:** the target's endurance costs in combat rise by `Value` percent. Reference.
- **Mesmerize duration buff:** lengthens the caster's mesmerizes by `Value` percent (Spiritmaster). Provisional: no source for the cap.

## 6.1 Settled

1. One mechanism: the debuffs are negative buffs on an effect, with the duration rule above; stat ones lower the stat, resist ones the resist, combat speed lengthens the swing, armour absorption lowers absorption.
2. Creatures cast their template's `StrengthBuff`, `StrengthConstitutionBuff`, `BaseArmorFactorBuff`, `StrengthDebuff` and `DexterityDebuff`; the census creature types reach 15 of 15.
3. Nearsight immunity after it ends: 10 seconds, provisional.
