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
