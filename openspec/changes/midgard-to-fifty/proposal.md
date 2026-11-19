## Why

After `thane-and-skald`, eight of Midgard's twelve classes play and the small-group PvE loop works. A 1.65 Midgard group would still notice a lot missing:

- **No resurrection.** A wipe means everyone releases.
- **No concentration.** Healer and Shaman buffs are flat and unlimited.
- **No line of sight.** Spells reach through dungeon walls and over hills.
- **No merchant gear past 30.** Merchants sell nothing above level 30 (the census shows 32 gaps).
- **Four classes missing.** The Spiritmaster, Bonedancer, Hunter and Shadowblade can't be made: they need pets, archery, stealth, poisons and left-hand attacks, none of which exist yet.
- **Nothing to do but kill.** There are no quests or tasks.

This change takes Midgard from "PvE to 30 with eight classes" to "PvE to 50 with all twelve".

## What Changes

The work is in four parts. Each part is shippable on its own and goes to the live server when it is done.

**Part A. Group essentials**
- **Line of sight.** The server decides line of sight from what it already loads: the terrain heightmap, building and dungeon hulls, and tree trunks. A player's spell, a creature's spell and (from part C) an arrow all need a clear line from caster to target, checked at the start and at the end of the cast. A caster who can't see the target gets the 1.65 message, and a creature that can't see its target moves closer instead of casting.
- **Resurrection.** The `Resurrect` spells of the Healer, Shaman and Spiritmaster (Seer and Mystic lines) target a dead player who has not released.
  - The raised player stands up with the spell's `ResurrectHealth` and `ResurrectMana` percent.
  - They regain part of the experience the death cost, and take on resurrection sickness.
  - The dead player is asked to accept, and releasing still works while they wait.
- **Concentration.** Every caster has a concentration pool that grows with level. A spell with `Concentration` > 0 is a concentration buff: it stays on its target until cancelled, until the caster is out of range, or until the caster dies, and its cost is held from the pool for as long as it is on. The buffs a Healer and a Shaman are known for come in with it:
  - the stat buffs (`ConstitutionBuff`, `DexterityBuff`, `AcuityBuff`) and the existing ones as concentration;
  - haste (`CombatSpeedBuff`) and `DamageShield`;
  - `EnduranceRegenBuff` and `PowerRegenBuff`;
  - the cures (`CurePoison`, `CureDisease`, `CureMezz`, `CureNearsightCustom`) and `SpreadHeal`.

  A concentration window lists what each caster holds.

**Part B. Levelling gear to 50**
- **Merchant stock.** Merchants stock the empty bands from 31 to 50. Each band's weapons, shields and armour come from the data's own non-unique Midgard item templates of that level (the data has 160 to 520 such templates in each band). They are sold in the towns near level-30-plus hunting: Jordheim and the frontier towns.
- **Sources first.** Which towns sold what, and at what quality, is looked up before the stock is chosen, and recorded with a status. Generic loot already reaches level 50 and is not changed. The census gap count goes to 0.

**Part C. Pets, then the last four classes**
- **Debuffs.** Strength, dexterity, strength/constitution and dexterity/quickness debuffs; combat-speed debuffs; resist debuffs; `Disease`, `Nearsight`, `Confusion` and `Amnesia`; `UnbreakableSpeedDecrease`; `DamageSpeedDecrease`; `DirectDamageWithDebuff`; and `FatigueConsumptionDebuff`. This also covers the creature spell types still missing (five of fifteen).
- **Pets.** A pet is a creature that follows its owner and fights what it is told to.
  - It is summoned from the spell's npc template (`LifeDrainReturn`), at a level set by the spell's `Value`.
  - It takes owner commands: attack, follow, stay, passive, defensive, aggressive, release. They are sent from a pet window and a `/pet` chat command.
  - It is part of the owner's fights: experience, aggro and con colours are reckoned with the owner.
  - It dies or leaves when its owner dies, zones or goes too far away.
- **Charm.** The Hunter charms an animal up to the spell's level. The charm can break.
- **Bonedancer.** A commander (`SummonCommander`) leads up to three minions (`SummonMinion`), and minion spells act through the commander.
- **Pet-caster spell types.** `Bladeturn`, `AblativeArmor`, `ArmorAbsorptionBuff`, `LifeTransfer`, `MesmerizeDurationBuff`.
- **Spiritmaster** (Summoning, Darkness, Suppression) and **Bonedancer** (Bone Army, Darkness, Suppression).
- **Hunter.**
  - Archery: composite bows, arrows as ammunition, draw and release, range and the line-of-sight check.
  - The bow abilities: Critical Shot, Rapid Fire and Penetrating Arrow.
  - The Beastcraft pet, Camouflage, and the Spear and Sword lines it already has.
- **Shadowblade.**
  - Stealth: hidden from creatures and players beyond a detection range that depends on level, Stealth spec and the viewer's Detect Hidden. It breaks on attacking, casting and taking damage. Stealth movement speed; the three stealth-opener styles; Danger Sense, Detect Hidden, Distraction, Safe Fall and Climb Walls as far as each applies to PvE.
  - Poisons: Envenom poison items applied to a weapon. They fire as weapon procs, with charges.
  - Left-hand attacks: the left axe swings alongside the main hand, under the left-hand rule. The Savage's hand-to-hand uses the same rule.
- **Trainers.** The four classes' trainers already stand in the world (8 placements each). They teach once the classes are in the class table.

**Part D. Tasks and quests**
- **Kill and delivery tasks.** Tasks come from Midgard town NPCs, as in classic Midgard.
  - A kill task names a creature of the player's con band in a nearby zone. A delivery task carries an item to another NPC.
  - The player holds one task at a time, under a time limit and a limit of tasks per level.
  - The reward is experience and money from a rule in the rule book.
- **The data's quests are not classic.** All 81 Midgard `dataquest` rows are "XP Item" turn-ins, a later freeshard addition. They are marked not classic and stay inert, like teleporters and hasteners. The census reports them in that column.

**Throughout**
- Each new mechanic gets a `rules-*` spec, a rule-book registry entry with sources and a status, and table tests.
- The importer gains, all additive:
  - spell columns: `concentration`, `resurrect_health`, `resurrect_mana`, `is_focus`;
  - npc template columns the pets need: stats, body type, spells, styles, left-hand chance;
  - a `dataquest` count;
  - poison items' charges.
- The protocol gains messages and fields, all additive: pet, concentration, resurrection offer, stealth visibility, ranged attack and task.
- `docs/parity`, the README, CLAUDE.md and the project status are updated at the end of each part.

**Not in scope**
- Realm-versus-realm: keeps, relics, battlegrounds, and the other realms.
- Crafting and housing.
- Realm abilities.
- The Mystic and Seer base-class gaps, beyond the spells their subclasses use.
- Scripted (non-data) 1.65 quests and epic quests. Their content isn't in the data and needs its own change and sources.

## Capabilities

### New Capabilities

- `rules-line-of-sight`: when a caster or archer can see a target, measured against terrain, hulls and trunks; what a blocked caster and a blocked creature do.
- `rules-concentration`: the concentration pool, which spells use it, how long a concentration buff lasts, range and death, and the concentration window.
- `pets`: summoning, owner commands, following, pet combat, charm and its breaking, Bonedancer minions, how pets share experience and aggro, and the pet window with its fallback.
- `rules-archery`: bows, arrows, draw time, ranged damage and hit, Critical Shot, Rapid Fire and Penetrating Arrow.
- `rules-stealth`: hiding, detection range, what breaks stealth, stealth movement, stealth openers, and the Shadowblade utility abilities.
- `rules-poisons`: applying a poison to a weapon, charges, and how a poison fires.
- `rules-tasks`: getting, holding, finishing and abandoning kill and delivery tasks, and their rewards.

### Modified Capabilities

- `midgard-classes`: Spiritmaster, Bonedancer, Hunter and Shadowblade become playable, so all twelve Midgard classes are offered.
- `rules-character-stats`: the class table gains the four classes, and casters gain a concentration pool.
- `player-spellcasting`: line-of-sight checks, resurrection, concentration buffs, summon and charm spells, debuffs, and the new buff and cure types.
- `rules-spell-effects`: how debuffs, haste, the damage shield, bladeturn, ablative armour, cures, the regeneration buffs and life transfer work.
- `rules-death`: resurrection, as an alternative to releasing.
- `rules-creature-casting`: creatures need line of sight, and cast the buff and debuff types they carry.
- `rules-melee`: left-hand attacks, and stealth-opener styles.
- `merchants`: stock for levels 31 to 50.
- `world-data-import`: the new spell, template and item columns, and data quests marked not classic.

## Impact

- `internal/server/rules`: new rule files for line of sight, concentration, pets, archery, stealth, poisons and tasks; `character.go` (four classes, concentration pool); `spells.go` (debuffs, buff types, resurrection); `melee.go` (left hand); `death.go` (resurrection experience). Table tests and registry entries for each.
- `internal/server/world`:
  - `collision.go` gains a segment test, used by `pcasting.go`, `casting.go` and archery;
  - `effects.go` gains debuffs, concentration buffs and cures;
  - new files: `pets.go`, `archery.go`, `stealth.go`, `tasks.go`;
  - `npc.go`: pet and charm brains, stealth detection;
  - `death.go`: the resurrection offer;
  - `shop.go` / catalog: the new merchant stock.
- `internal/importer`: additive columns, plus `notclassic.go` for data quests. `world.db` is re-imported in each part that changes it.
- `web/`:
  - windows: the pet window, the concentration list, the resurrection prompt and the task window, each with a plain fallback;
  - stealth drawing (translucent self, unseen others);
  - the ranged attack and draw bar;
  - `move.js`: stealth speed through a shared vector.
- `docs/protocol.md`: additive messages and fields, with protocol tests.
- Converted creature models: each pet's model must be converted with `charconv` (`web/data/char/m<model>`), like other creatures. This adds to the first-run recipe.
- `cmd/parity`, `docs/parity`, the README, CLAUDE.md ("Open now").
- Live server: for each part, back up `players.db`, re-import, rebuild and restart. Existing characters are unchanged.
