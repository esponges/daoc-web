# Parity census

How much of classic 1.65 Midgard the server implements, measured against the community database by `go run ./cmd/parity`. Counts and names only; the rules are in `openspec/specs/rules-*` and `internal/server/rules`.

## Summary

```
Midgard (regions 100, 101, 151): 29932 placements
  Creature spell types    15 of  15 implemented (2585 of 2585 placements x spells)
  Class spell types       64 of  73 implemented (939 of 987 spells)
  Style features          12 of  14 implemented (1379 of 1383 styles)
  Abilities               18 of  30 implemented (80 of 100 grants)
  Creature kinds          22 of 110 implemented (29110 of 29932 placements, 7 kinds not classic)
  quests                 81, none implemented
  keep doors             432, none implemented
  keeps                  79, none implemented
  relics                 6, none implemented
  battlegrounds          5, none implemented
  levelling gear         0 gaps in 10 bands
  rules: 26 documented, 64 reference, 31 provisional
```

## Creature spell types

| name | placements x spells | implemented |
| --- | --- | --- |
| DirectDamage | 744 | yes |
| SpeedDecrease | 372 | yes |
| OffensiveProc | 334 | yes |
| DamageOverTime | 307 | yes |
| Mesmerize | 300 | yes |
| BaseArmorFactorBuff | 155 | yes |
| Bolt | 133 | yes |
| Heal | 46 | yes |
| StrengthConstitutionBuff | 44 | yes |
| StrengthBuff | 38 | yes |
| Lifedrain | 37 | yes |
| Stun | 37 | yes |
| StrengthDebuff | 21 | yes |
| DexterityDebuff | 11 | yes |
| DirectDamageNoVariance | 6 | yes |

## Class spell types

| name | spells | implemented |
| --- | --- | --- |
| DirectDamage (Bonedancer, Mystic, Runemaster, Shaman, Skald, Spiritmaster, Thane) | 100 | yes |
| Heal (Bonedancer, Healer, Seer, Shaman, Spiritmaster) | 75 | yes |
| DamageOverTime (Bonedancer, Shadowblade, Shaman) | 56 | yes |
| DamageAdd (Bonedancer, Healer, Mystic, Runemaster, Seer, Shaman, Skald, Spiritmaster, Thane) | 46 | yes |
| Mesmerize (Bonedancer, Healer, Mystic, Runemaster, Shadowblade, Skald, Spiritmaster) | 44 | yes |
| SpeedDecrease (Bonedancer, Healer, Mystic, Runemaster, Shadowblade, Shaman, Skald, Spiritmaster) | 40 | yes |
| StrengthConstitutionBuff (Bonedancer, Healer, Hunter, Mystic, Runemaster, Seer, Shaman, Spiritmaster, Thane) | 40 | yes |
| SummonMinion (Bonedancer, Mystic, Runemaster, Spiritmaster) | 36 | yes |
| DexterityQuicknessBuff (Bonedancer, Healer, Hunter, Seer, Shaman, Spiritmaster) | 30 | yes |
| Bolt (Runemaster, Shaman) | 29 | yes |
| BaseArmorFactorBuff (Bonedancer, Healer, Hunter, Mystic, Runemaster, Seer, Shaman, Spiritmaster) | 26 | yes |
| DamageShield (Healer, Seer, Shaman, Spiritmaster) | 26 | yes |
| DamageSpeedDecrease (Bonedancer, Mystic, Runemaster, Spiritmaster) | 24 | yes |
| Lifedrain (Bonedancer, Mystic, Runemaster, Spiritmaster) | 21 | yes |
| SpeedEnhancement (Healer, Hunter, Runemaster, Seer, Shaman, Skald) | 20 | yes |
| CombatSpeedBuff (Healer, Seer, Shaman, Spiritmaster) | 16 | yes |
| CombatSpeedDebuff (Bonedancer, Healer, Mystic, Runemaster, Shadowblade, Spiritmaster) | 16 | yes |
| HealthRegenBuff (Healer, Seer, Shaman, Skald) | 16 | yes |
| DirectDamageWithDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 15 | yes |
| Stun (Healer) | 14 | yes |
| Disease (Shadowblade, Shaman) | 13 | yes |
| OffensiveProc (Shadowblade, Spiritmaster) | 13 | yes |
| StrengthConstitutionDebuff (Bonedancer, Mystic, Runemaster, Shadowblade, Spiritmaster) | 12 | yes |
| StrengthDebuff (Bonedancer, Mystic, Runemaster, Shadowblade, Spiritmaster) | 12 | yes |
| Charm (Hunter) | 10 | yes |
| Bladeturn (Bonedancer, Mystic, Runemaster, Spiritmaster) | 8 | yes |
| DexterityDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 8 | yes |
| StrengthBuff (Healer, Seer, Shaman) | 8 | yes |
| UnbreakableSpeedDecrease (Shadowblade) | 8 | yes |
| Confusion (Bonedancer, Mystic, Runemaster, Spiritmaster) | 7 | yes |
| ConstitutionBuff (Healer, Seer, Shaman) | 7 | yes |
| DexterityQuicknessDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 7 | yes |
| LifeTransfer (Bonedancer, Mystic, Runemaster, Spiritmaster) | 7 | yes |
| AblativeArmor (Bonedancer, Mystic, Runemaster, Spiritmaster) | 6 | yes |
| BodyResistDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 6 | yes |
| DexterityBuff (Healer, Seer, Shaman) | 6 | yes |
| Resurrect (Healer, Seer, Shaman, Spiritmaster) | 6 | yes |
| SavageCrushResistanceBuff (Savage) | 6 |  |
| SavageDPSBuff (Savage) | 6 |  |
| SavageEvadeBuff (Savage) | 6 |  |
| SavageParryBuff (Savage) | 6 |  |
| SavageSlashResistanceBuff (Savage) | 6 |  |
| SavageThrustResistanceBuff (Savage) | 6 |  |
| Amnesia (Healer) | 5 | yes |
| ArmorAbsorptionBuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 5 | yes |
| BodyResistBuff (Healer, Seer, Shaman, Skald) | 5 | yes |
| ColdResistBuff (Healer, Seer, Shaman, Skald) | 5 | yes |
| EnduranceRegenBuff (Healer, Seer, Shaman) | 5 | yes |
| EnergyResistBuff (Healer, Seer, Shaman, Skald) | 5 | yes |
| HeatResistBuff (Healer, Seer, Shaman, Skald) | 5 | yes |
| MatterResistBuff (Healer, Seer, Shaman, Skald) | 5 | yes |
| Nearsight (Bonedancer, Mystic, Runemaster, Spiritmaster) | 5 | yes |
| PowerRegenBuff (Healer) | 5 | yes |
| SavageCombatSpeedBuff (Savage) | 5 |  |
| SpiritResistBuff (Healer, Seer, Shaman, Skald) | 5 | yes |
| SummonHunterPet (Hunter) | 5 | yes |
| SummonSpiritFighter (Spiritmaster) | 5 | yes |
| AcuityBuff (Healer, Seer, Shaman) | 4 | yes |
| CureNearsightCustom (Healer, Seer, Shaman) | 4 | yes |
| SummonCommander (Bonedancer) | 4 | yes |
| Taunt (Savage) | 4 |  |
| ColdResistDebuff (Runemaster) | 3 | yes |
| EnergyResistDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 3 | yes |
| HeatResistDebuff (Runemaster) | 3 | yes |
| MatterResistDebuff (Runemaster) | 3 | yes |
| SavageEnduranceHeal (Savage) | 3 |  |
| SpiritResistDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 3 | yes |
| SpreadHeal (Healer, Seer, Shaman) | 3 | yes |
| CureDisease (Healer, Seer, Shaman) | 2 | yes |
| CureMezz (Bonedancer, Healer, Mystic, Runemaster, Spiritmaster) | 2 | yes |
| CurePoison (Healer, Seer, Shaman) | 2 | yes |
| FatigueConsumptionDebuff (Shadowblade) | 2 | yes |
| MesmerizeDurationBuff (Spiritmaster) | 2 | yes |

## Style features

| name | styles | implemented |
| --- | --- | --- |
| growth damage | 341 | yes |
| to-hit bonus | 280 | yes |
| proc spell | 264 | yes |
| defence bonus | 132 | yes |
| chain: after a style | 130 | yes |
| opening: anytime | 102 | yes |
| opening: from behind | 28 | yes |
| opening: after a block | 26 | yes |
| opening: after a parry | 26 | yes |
| opening: from the side | 25 | yes |
| opening: after an evade | 18 | yes |
| opening: from the front | 7 | yes |
| stealth opener | 3 |  |
| opening: after result 3 | 1 |  |

## Abilities

| name | grants | implemented |
| --- | --- | --- |
| Evade (Berserker, Hunter, Midgard Rogue, Savage, Shadowblade, Skald, Warrior) | 17 | yes |
| MidArmor (Berserker, Bonedancer, Healer, Hunter, Midgard Rogue, Mystic, Runemaster, Savage, Seer, Shadowblade, Shaman, Skald, Spiritmaster, Thane, Viking, Warrior) | 10 | yes |
| Critical Shot (Hunter) | 9 | yes |
| Protect (Berserker, Skald, Thane, Warrior) | 8 | yes |
| Shield (Berserker, Healer, Savage, Seer, Shadowblade, Shaman, Skald, Thane, Viking, Warrior) | 8 | yes |
| Safe Fall (Shadowblade) | 5 |  |
| Berserk (Berserker) | 4 |  |
| Guard (Thane, Warrior) | 4 | yes |
| Weaponry: Staves (Berserker, Bonedancer, Healer, Hunter, Midgard Rogue, Mystic, Runemaster, Savage, Seer, Shadowblade, Shaman, Skald, Spiritmaster, Thane, Viking, Warrior) | 4 | yes |
| Penetrating Arrow (Hunter) | 3 | yes |
| Weaponry: Thrown (Berserker, Shadowblade, Warrior) | 3 | yes |
| Charge (Berserker, Savage) | 2 |  |
| Intercept (Thane, Warrior) | 2 | yes |
| Rapid Fire (Hunter) | 2 | yes |
| Weaponry: Axes (Berserker, Savage, Shadowblade, Skald, Thane, Viking, Warrior) | 2 | yes |
| Weaponry: Hammers (Berserker, Healer, Savage, Seer, Shaman, Skald, Thane, Viking, Warrior) | 2 | yes |
| Weaponry: Swords (Berserker, Hunter, Midgard Rogue, Savage, Shadowblade, Skald, Thane, Viking, Warrior) | 2 | yes |
| Call of a thousand storms (Thane) | 1 |  |
| Camouflage (Hunter) | 1 |  |
| Climb Walls (Shadowblade) | 1 |  |
| Danger Sense (Shadowblade) | 1 |  |
| Detect Hidden (Shadowblade) | 1 |  |
| Distraction (Shadowblade) | 1 |  |
| Engage (Thane, Warrior) | 1 | yes |
| Enhanced Evade (Berserker, Savage) | 1 |  |
| QuickCast (Bonedancer, Runemaster, Spiritmaster) | 1 |  |
| Weaponry: Composite Bows (Hunter) | 1 | yes |
| Weaponry: Hand to Hand (Savage) | 1 |  |
| Weaponry: Left Axes (Shadowblade) | 1 | yes |
| Weaponry: Spears (Hunter) | 1 | yes |

## Creature kinds

| name | placements | implemented |
| --- | --- | --- |
| DOL.GS.GameNPC | 28412 | yes |
| DOL.GS.GameMerchant | 324 | yes |
| DOL.GS.GameGuard | 197 | yes |
| DOL.GS.Scripts.Myling | 194 |  |
| DOL.GS.Keeps.GuardFighter | 139 |  |
| DOL.GS.NightSpawn | 71 |  |
| DOL.GS.Scripts.ShiveringPresence | 51 |  |
| DOL.GS.Keeps.GuardArcher | 49 |  |
| DOL.GS.GameHastener | 27 | not classic: a hastener: a speed boost on request |
| DOL.GS.Keeps.GuardCommander | 26 |  |
| DOL.GS.Blacksmith | 23 | yes |
| DOL.GS.GameStableMaster | 23 | yes |
| DOL.GS.QueenMajorAdd | 22 |  |
| DOL.GS.GameHealer | 18 | yes |
| DOL.GS.Scripts.OFAssistant | 18 | not classic: a teleporter's assistant |
| DOL.GS.Scripts.LiveTeleporter | 15 | not classic: a teleporter: Live-era travel the horse routes replace |
| DOL.GS.Enchanter | 14 |  |
| DOL.GS.StripeAdd | 13 |  |
| DOL.GS.ZritZritAdd | 13 |  |
| DOL.GS.Keeps.FrontierHastener | 11 | not classic: a hastener: a speed boost on request |
| DOL.GS.GameVaultKeeper | 9 | yes |
| DOL.GS.OonaUndeadSoldier | 9 |  |
| DOL.GS.Keeps.GuardStaticArcher | 8 |  |
| DOL.GS.RainbowSpriteWhite | 8 |  |
| DOL.GS.Trainer.BerserkerTrainer | 8 | yes |
| DOL.GS.Trainer.BonedancerTrainer | 8 | yes |
| DOL.GS.Trainer.HealerTrainer | 8 | yes |
| DOL.GS.Trainer.HunterTrainer | 8 | yes |
| DOL.GS.Trainer.RunemasterTrainer | 8 | yes |
| DOL.GS.Trainer.ShadowbladeTrainer | 8 |  |
| DOL.GS.Trainer.ShamanTrainer | 8 | yes |
| DOL.GS.Trainer.SkaldTrainer | 8 | yes |
| DOL.GS.Trainer.SpiritmasterTrainer | 8 | yes |
| DOL.GS.Trainer.ThaneTrainer | 8 | yes |
| DOL.GS.Trainer.WarriorTrainer | 8 | yes |
| DOL.GS.Keeps.GuardCorpseSummoner | 7 |  |
| DOL.GS.Keeps.GuardLord | 7 |  |
| DOL.GS.RainbowSpriteBlue | 7 |  |
| DOL.GS.RainbowSpriteTan | 7 |  |
| DOL.GS.Trainer.SavageTrainer | 7 | yes |
| DOL.GS.RainbowSpriteGreen | 6 |  |
| DOL.GS.Recharger | 5 |  |
| DOL.GS.Trainer.SeerTrainer | 5 | yes |
| DOL.GS.Trainer.VikingTrainer | 5 | yes |
| DOL.GS.DPSDummy | 4 |  |
| DOL.GS.Trainer.MysticTrainer | 4 | yes |
| DOL.GS.GameBountyMerchant | 3 | not classic: a bounty merchant: a later reward shop |
| DOL.GS.MidgardSITeleporter | 3 |  |
| DOL.GS.Scripts.OFTeleporter | 3 | not classic: a teleporter: Origin Fantasy convenience travel |
| DOL.GS.SnarlsAdd | 3 |  |
| DOL.GS.Trainer.RogueTrainer | 3 | yes |
| DOL.GS.AlchemistsMaster | 2 |  |
| DOL.GS.ArmorCraftingMaster | 2 |  |
| DOL.GS.EmblemNPC | 2 |  |
| DOL.GS.FletchingMaster | 2 |  |
| DOL.GS.GuildRegistrar | 2 |  |
| DOL.GS.HitbackDummy | 2 |  |
| DOL.GS.Keeps.GuardMerchant | 2 |  |
| DOL.GS.Loken | 2 |  |
| DOL.GS.Scripts.OFMerchantHome | 2 |  |
| DOL.GS.Scripts.Recharger | 2 |  |
| DOL.GS.TailoringMaster | 2 |  |
| DOL.GS.WeaponCraftingMaster | 2 |  |
| DOL.GS.AccountVaultKeeper | 1 | not classic: an account vault: shared storage across characters |
| DOL.GS.Alina | 1 |  |
| DOL.GS.Basilius | 1 |  |
| DOL.GS.Boligar | 1 |  |
| DOL.GS.ChiefTarek | 1 |  |
| DOL.GS.Cronk | 1 |  |
| DOL.GS.Dalgo | 1 |  |
| DOL.GS.DaySpawn | 1 |  |
| DOL.GS.EffectNPC | 1 |  |
| DOL.GS.GiantLemer | 1 |  |
| DOL.GS.GlacierGiant | 1 |  |
| DOL.GS.Goretooth | 1 |  |
| DOL.GS.HealDummy | 1 |  |
| DOL.GS.Hugi | 1 |  |
| DOL.GS.MidGjalpinulva | 1 |  |
| DOL.GS.Mortufoghus | 1 |  |
| DOL.GS.NameRegistrar | 1 |  |
| DOL.GS.Njessi | 1 |  |
| DOL.GS.Oona | 1 |  |
| DOL.GS.QueenMajor | 1 |  |
| DOL.GS.RaumarikRevenant | 1 |  |
| DOL.GS.RoyalTreasuryClerk | 1 |  |
| DOL.GS.Runil | 1 |  |
| DOL.GS.ScoutArgyle | 1 |  |
| DOL.GS.Scripts.AtlasTrainer | 1 |  |
| DOL.GS.Scripts.InlandTeleporter | 1 |  |
| DOL.GS.Shredclaw | 1 |  |
| DOL.GS.SimpleTeleporter | 1 |  |
| DOL.GS.Snarls | 1 |  |
| DOL.GS.SpellCraftingMaster | 1 |  |
| DOL.GS.Steinolf | 1 |  |
| DOL.GS.Stripe | 1 |  |
| DOL.GS.Trainer.MidgardRogueTrainer | 1 |  |
| DOL.GS.Ulfketill | 1 |  |
| DOL.GS.Valmak | 1 |  |
| DOL.GS.Vasrem | 1 |  |
| DOL.GS.WrethOfWinterArachite | 1 |  |
| DOL.GS.WrethOfWinterCrab | 1 |  |
| DOL.GS.WrethOfWinterIce | 1 |  |
| DOL.GS.WrethOfWinterJotun | 1 |  |
| DOL.GS.WrethOfWinterLich | 1 |  |
| DOL.GS.WrethOfWinterOgre | 1 |  |
| DOL.GS.WrethOfWinterRaumarik | 1 |  |
| DOL.GS.WrethOfWinterScorp | 1 |  |
| DOL.GS.WrethOfWinterSpider | 1 |  |
| DOL.GS.Yar | 1 |  |
| DOL.GS.ZritZrit | 1 |  |

## Levelling gear

Midgard merchants (placements) selling Warrior gear in each five-level band. **none** marks a gap: a band where a Warrior needs that kind and no merchant sells it. Gaps are reported, not filled.

| band | 1h axe | 1h hammer | 1h sword | 2h axe | 2h hammer | 2h sword | shield | studded | chain |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1-5 | 17 | 9 | 21 | 11 | 11 | 11 | 13 | 26 | 19 |
| 6-10 | 17 | 9 | 21 | 11 | 11 | 11 | 13 | 26 | 22 |
| 11-15 | 17 | 9 | 21 | 11 | 11 | 11 | 13 | 26 | 22 |
| 16-20 | 17 | 9 | 21 | 11 | 11 | 11 | 13 | 26 | 22 |
| 21-25 | 17 | 9 | 21 | 11 | 11 | 11 | 13 | 22 | 15 |
| 26-30 | 7 | 7 | 7 | 9 | 9 | 9 | 13 | 5 | 6 |
| 31-35 | 2* | 2* | 2* | 2* | 2* | 2* | 8* | 5* | 1* |
| 36-40 | 2* | 2* | 2* | 2* | 2* | 2* | 8* | 5* | 1* |
| 41-45 | 2* | 2* | 2* | 2* | 2* | 2* | 8* | 5* | 1* |
| 46-50 | 2* | 2* | 2* | 2* | 2* | 2* | 8* | 5* | 1* |

## Rules

| rule | status |
| --- | --- |
| ability.engage | reference |
| ability.gained | documented |
| ability.guard | reference |
| ability.intercept | reference |
| ability.protect | reference |
| ability.target | reference |
| archery.abilities | provisional |
| archery.damage | provisional |
| archery.draw | provisional |
| archery.range | provisional |
| cast.decision | provisional |
| cast.heal | provisional |
| cast.interrupt | reference |
| cast.time | reference |
| character.class | reference |
| character.class-hunter | provisional |
| character.class-midgard | provisional |
| character.class-spiritmaster-bonedancer | provisional |
| character.class-thane-skald | provisional |
| character.creation | documented |
| character.endurance | provisional |
| character.health-regen | provisional |
| character.hit-points | reference |
| character.item-stats | documented |
| character.level-stats | documented |
| character.race-stats | documented |
| conc.end | provisional |
| conc.hold | reference |
| conc.pool | reference |
| condition.effectiveness | reference |
| condition.repair-price | provisional |
| condition.start | reference |
| condition.warning | provisional |
| condition.wear | provisional |
| creature.armour | reference |
| creature.blows | reference |
| creature.defences | documented |
| creature.hit-points | reference |
| creature.level-one | reference |
| creature.resists | documented |
| creature.stats | reference |
| creature.swing | reference |
| death.constitution | reference |
| death.healer | documented |
| death.loss | reference |
| death.pray | reference |
| death.release | reference |
| death.return | reference |
| experience.con | documented |
| experience.creature-value | documented |
| experience.group-grey | provisional |
| experience.group-share | provisional |
| experience.group-split | provisional |
| experience.kill | documented |
| experience.table | documented |
| loot.generic | provisional |
| loot.money | provisional |
| los.cast | reference |
| los.cost | provisional |
| los.creature | reference |
| los.visible | provisional |
| melee.armour | documented |
| melee.base-damage | documented |
| melee.block | documented |
| melee.creature-proc | reference |
| melee.critical | reference |
| melee.evade | documented |
| melee.location | reference |
| melee.miss | documented |
| melee.miss-bonuses | reference |
| melee.order | reference |
| melee.parry | documented |
| melee.style-cost | reference |
| melee.style-damage | documented |
| melee.style-spell | reference |
| melee.swing-speed | documented |
| melee.weapon-proc | reference |
| melee.weapon-skill | reference |
| pet.charm | provisional |
| pet.level | reference |
| pet.minions | reference |
| power.cost | reference |
| power.pool | reference |
| power.regen | provisional |
| rez.sickness | reference |
| rez.spare | documented |
| rez.wait | reference |
| spec.abilities | documented |
| spec.items | documented |
| spec.migration | documented |
| spec.points | reference |
| spec.training | documented |
| spell.area | provisional |
| spell.bolt-block | reference |
| spell.buff | reference |
| spell.buff-af | provisional |
| spell.buff-cda | reference |
| spell.control-debuffs | reference |
| spell.damage | reference |
| spell.damage-add | reference |
| spell.damage-over-time | reference |
| spell.debuffs | reference |
| spell.haste | provisional |
| spell.heal | reference |
| spell.health-regen | provisional |
| spell.hit | reference |
| spell.life-transfer | reference |
| spell.lifedrain | reference |
| spell.mesmerize | reference |
| spell.mez-duration | provisional |
| spell.regen-buffs | reference |
| spell.resist | reference |
| spell.resist-buff | reference |
| spell.slow | reference |
| spell.speed-enhancement | reference |
| spell.stun | reference |
| spell.wards | reference |
| start.town | documented |
| start.unsafe | provisional |
| time.day-length | reference |
| time.night | reference |
