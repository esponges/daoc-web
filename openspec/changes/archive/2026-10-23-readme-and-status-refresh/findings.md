# What blocks a friends group from levelling 1 to 50

Read from `docs/parity/README.md` (the census, regenerated on 2026-10-22) and the server
code. "Playable" means the creation page offers the class: Warrior, Berserker, Savage,
Healer, Shaman, Runemaster. Midgard had twelve classes in 1.65 (`cmd/parity/parity.go`,
`midgardClasses`, minus the base classes Viking, Mystic, Seer and Rogue).

Corrections to what was said while exploring: trainers exist (`world/services.go`
`openTrainer`, spec `rules-specialization`), and Midgard has twelve classes, not ~20.

## Ranked

1. **Gear above level 30.** The levelling-gear table (`docs/parity/README.md`, "Levelling
   gear") shows every weapon, shield and chain row as **none** for bands 31-35 to 46-50
   (32 gaps in 10 bands); the 26-30 band already thins from 17 to 7 weapon sellers.
   Merchants are the only source this check counts; whether drops cover the gap is
   not measured. This is the first thing a character meets after level 30.
   Source: census `levellingGear`.
2. **The six unplayable classes.** Thane, Skald, Hunter, Shadowblade, Spiritmaster and
   Bonedancer have trainers in the data (`DOL.GS.Trainer.*Trainer`, 8 placements
   each) but no server class. Two of them need systems that do not exist:
   Spiritmaster and Bonedancer need pets (`SummonMinion`, 36 spells, not
   implemented), Shadowblade needs stealth (3 stealth-opener styles and Camouflage,
   Danger Sense, Detect Hidden, Climb Walls abilities not implemented).
   Source: census "Abilities", "Creature kinds", "Class spell types".
3. **Class spells: 13 of 73 types (529 of 987 spells).** The big unimplemented ones
   are `SummonMinion` 36, `DamageShield` 26, `DamageSpeedDecrease` 24,
   `SpeedEnhancement` 20, and the combat-speed and regeneration buffs. Healer,
   Shaman and Runemaster lines are the ones that matter for the six playable
   classes. Source: census "Class spell types".
4. **Abilities of the playable classes.** Berserk (Berserker, 4 grants), Charge
   (Berserker, Savage, 2), Enhanced Evade (1), Weaponry: Hand to Hand (Savage, 1) are
   not implemented. 15 of 30 overall. Source: census "Abilities".
5. **Creature coverage at high level.** 16 of 110 kinds are implemented, which covers
   29,067 of 29,932 placements (97%). The 865 left include Myling (194), Keeps guards
   (139 + 49 + 26), NightSpawn (71), Shivering Presence (51) and raid adds. Five
   creature spell types (armour factor, strength and dexterity buffs and debuffs, 269
   placement-spell uses) are also missing, which matters for dungeon mobs.
   Source: census "Creature kinds", "Creature spell types".
6. **Quests: 0 of 81.** Not needed to reach 50 in 1.65, which levelled by killing, but the
   journal button is greyed for this reason (`web/chrome.js` `MENU_WHY`).
7. **RvR: keeps 0/79, keep doors 0/432, relics 0/6, battlegrounds 0/5.** Last, and
   out of reach for a friends group until the PvE path is whole. Also pets, charm,
   line of sight, concentration, crafting and housing, none of which block levelling.

## Suggested follow-up change

`path-to-fifty`, in two parts, because 1 to 3 are independent of 4 to 5:

- Gear and class parity: fill gear for 31-50 (decide first whether from merchants,
  drops, or both), then the remaining abilities of the playable six, then the spell
  types those six use.
- Reach and breadth: the next classes that need no new system (Thane, Skald),
  then Hunter (ranged combat), and last the two that need pets or stealth.

An open question for the user, not decided here: whether the target is "every Midgard
class" or "enough classes for the friends who will play". The ranking above holds either way.
