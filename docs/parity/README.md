# Parity census

How much of classic 1.65 Midgard the server implements, measured against the community database by `go run ./cmd/parity`. Counts and names only; the rules are in `openspec/specs/rules-*` and `internal/server/rules`.

## Summary

```
Midgard (regions 100, 101, 151): 29932 placements
  Creature spell types     0 of  15 implemented (0 of 2585 placements x spells)
  Class spell types        0 of  73 implemented (0 of 987 spells)
  Style features          11 of  14 implemented (1115 of 1124 styles)
  Abilities               11 of  30 implemented (51 of 100 grants)
  Creature kinds           5 of 110 implemented (28767 of 29932 placements)
  quests                 81, none implemented
  keep doors             432, none implemented
  keeps                  79, none implemented
  relics                 6, none implemented
  battlegrounds          5, none implemented
  levelling gear         32 gaps in 10 bands
  rules: 24 documented, 20 reference, 4 provisional
```

## Creature spell types

| name | placements x spells | implemented |
| --- | --- | --- |
| DirectDamage | 744 |  |
| SpeedDecrease | 372 |  |
| OffensiveProc | 334 |  |
| DamageOverTime | 307 |  |
| Mesmerize | 300 |  |
| BaseArmorFactorBuff | 155 |  |
| Bolt | 133 |  |
| Heal | 46 |  |
| StrengthConstitutionBuff | 44 |  |
| StrengthBuff | 38 |  |
| Lifedrain | 37 |  |
| Stun | 37 |  |
| StrengthDebuff | 21 |  |
| DexterityDebuff | 11 |  |
| DirectDamageNoVariance | 6 |  |

## Class spell types

| name | spells | implemented |
| --- | --- | --- |
| DirectDamage (Bonedancer, Mystic, Runemaster, Shaman, Skald, Spiritmaster, Thane) | 100 |  |
| Heal (Bonedancer, Healer, Seer, Shaman, Spiritmaster) | 75 |  |
| DamageOverTime (Bonedancer, Shadowblade, Shaman) | 56 |  |
| DamageAdd (Bonedancer, Healer, Mystic, Runemaster, Seer, Shaman, Skald, Spiritmaster, Thane) | 46 |  |
| Mesmerize (Bonedancer, Healer, Mystic, Runemaster, Shadowblade, Skald, Spiritmaster) | 44 |  |
| SpeedDecrease (Bonedancer, Healer, Mystic, Runemaster, Shadowblade, Shaman, Skald, Spiritmaster) | 40 |  |
| StrengthConstitutionBuff (Bonedancer, Healer, Hunter, Mystic, Runemaster, Seer, Shaman, Spiritmaster, Thane) | 40 |  |
| SummonMinion (Bonedancer, Mystic, Runemaster, Spiritmaster) | 36 |  |
| DexterityQuicknessBuff (Bonedancer, Healer, Hunter, Seer, Shaman, Spiritmaster) | 30 |  |
| Bolt (Runemaster, Shaman) | 29 |  |
| BaseArmorFactorBuff (Bonedancer, Healer, Hunter, Mystic, Runemaster, Seer, Shaman, Spiritmaster) | 26 |  |
| DamageShield (Healer, Seer, Shaman, Spiritmaster) | 26 |  |
| DamageSpeedDecrease (Bonedancer, Mystic, Runemaster, Spiritmaster) | 24 |  |
| Lifedrain (Bonedancer, Mystic, Runemaster, Spiritmaster) | 21 |  |
| SpeedEnhancement (Healer, Hunter, Runemaster, Seer, Shaman, Skald) | 20 |  |
| CombatSpeedBuff (Healer, Seer, Shaman, Spiritmaster) | 16 |  |
| CombatSpeedDebuff (Bonedancer, Healer, Mystic, Runemaster, Shadowblade, Spiritmaster) | 16 |  |
| HealthRegenBuff (Healer, Seer, Shaman, Skald) | 16 |  |
| DirectDamageWithDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 15 |  |
| Stun (Healer) | 14 |  |
| Disease (Shadowblade, Shaman) | 13 |  |
| OffensiveProc (Shadowblade, Spiritmaster) | 13 |  |
| StrengthConstitutionDebuff (Bonedancer, Mystic, Runemaster, Shadowblade, Spiritmaster) | 12 |  |
| StrengthDebuff (Bonedancer, Mystic, Runemaster, Shadowblade, Spiritmaster) | 12 |  |
| Charm (Hunter) | 10 |  |
| Bladeturn (Bonedancer, Mystic, Runemaster, Spiritmaster) | 8 |  |
| DexterityDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 8 |  |
| StrengthBuff (Healer, Seer, Shaman) | 8 |  |
| UnbreakableSpeedDecrease (Shadowblade) | 8 |  |
| Confusion (Bonedancer, Mystic, Runemaster, Spiritmaster) | 7 |  |
| ConstitutionBuff (Healer, Seer, Shaman) | 7 |  |
| DexterityQuicknessDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 7 |  |
| LifeTransfer (Bonedancer, Mystic, Runemaster, Spiritmaster) | 7 |  |
| AblativeArmor (Bonedancer, Mystic, Runemaster, Spiritmaster) | 6 |  |
| BodyResistDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 6 |  |
| DexterityBuff (Healer, Seer, Shaman) | 6 |  |
| Resurrect (Healer, Seer, Shaman, Spiritmaster) | 6 |  |
| SavageCrushResistanceBuff (Savage) | 6 |  |
| SavageDPSBuff (Savage) | 6 |  |
| SavageEvadeBuff (Savage) | 6 |  |
| SavageParryBuff (Savage) | 6 |  |
| SavageSlashResistanceBuff (Savage) | 6 |  |
| SavageThrustResistanceBuff (Savage) | 6 |  |
| Amnesia (Healer) | 5 |  |
| ArmorAbsorptionBuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 5 |  |
| BodyResistBuff (Healer, Seer, Shaman, Skald) | 5 |  |
| ColdResistBuff (Healer, Seer, Shaman, Skald) | 5 |  |
| EnduranceRegenBuff (Healer, Seer, Shaman) | 5 |  |
| EnergyResistBuff (Healer, Seer, Shaman, Skald) | 5 |  |
| HeatResistBuff (Healer, Seer, Shaman, Skald) | 5 |  |
| MatterResistBuff (Healer, Seer, Shaman, Skald) | 5 |  |
| Nearsight (Bonedancer, Mystic, Runemaster, Spiritmaster) | 5 |  |
| PowerRegenBuff (Healer) | 5 |  |
| SavageCombatSpeedBuff (Savage) | 5 |  |
| SpiritResistBuff (Healer, Seer, Shaman, Skald) | 5 |  |
| SummonHunterPet (Hunter) | 5 |  |
| SummonSpiritFighter (Spiritmaster) | 5 |  |
| AcuityBuff (Healer, Seer, Shaman) | 4 |  |
| CureNearsightCustom (Healer, Seer, Shaman) | 4 |  |
| SummonCommander (Bonedancer) | 4 |  |
| Taunt (Savage) | 4 |  |
| ColdResistDebuff (Runemaster) | 3 |  |
| EnergyResistDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 3 |  |
| HeatResistDebuff (Runemaster) | 3 |  |
| MatterResistDebuff (Runemaster) | 3 |  |
| SavageEnduranceHeal (Savage) | 3 |  |
| SpiritResistDebuff (Bonedancer, Mystic, Runemaster, Spiritmaster) | 3 |  |
| SpreadHeal (Healer, Seer, Shaman) | 3 |  |
| CureDisease (Healer, Seer, Shaman) | 2 |  |
| CureMezz (Bonedancer, Healer, Mystic, Runemaster, Spiritmaster) | 2 |  |
| CurePoison (Healer, Seer, Shaman) | 2 |  |
| FatigueConsumptionDebuff (Shadowblade) | 2 |  |
| MesmerizeDurationBuff (Spiritmaster) | 2 |  |

## Style features

| name | styles | implemented |
| --- | --- | --- |
| growth damage | 341 | yes |
| to-hit bonus | 280 | yes |
| defence bonus | 132 | yes |
| chain: after a style | 130 | yes |
| opening: anytime | 102 | yes |
| opening: from behind | 28 | yes |
| opening: after a block | 26 | yes |
| opening: after a parry | 26 | yes |
| opening: from the side | 25 | yes |
| opening: after an evade | 18 | yes |
| opening: from the front | 7 | yes |
| proc spell | 5 |  |
| stealth opener | 3 |  |
| opening: after result 3 | 1 |  |

## Abilities

| name | grants | implemented |
| --- | --- | --- |
| Evade (Berserker, Hunter, Midgard Rogue, Savage, Shadowblade, Skald, Warrior) | 17 | yes |
| MidArmor (Berserker, Bonedancer, Healer, Hunter, Midgard Rogue, Mystic, Runemaster, Savage, Seer, Shadowblade, Shaman, Skald, Spiritmaster, Thane, Viking, Warrior) | 10 | yes |
| Critical Shot (Hunter) | 9 |  |
| Protect (Berserker, Skald, Thane, Warrior) | 8 |  |
| Shield (Berserker, Healer, Savage, Seer, Shadowblade, Shaman, Skald, Thane, Viking, Warrior) | 8 | yes |
| Safe Fall (Shadowblade) | 5 |  |
| Berserk (Berserker) | 4 |  |
| Guard (Thane, Warrior) | 4 |  |
| Weaponry: Staves (Berserker, Bonedancer, Healer, Hunter, Midgard Rogue, Mystic, Runemaster, Savage, Seer, Shadowblade, Shaman, Skald, Spiritmaster, Thane, Viking, Warrior) | 4 | yes |
| Penetrating Arrow (Hunter) | 3 |  |
| Weaponry: Thrown (Berserker, Shadowblade, Warrior) | 3 | yes |
| Charge (Berserker, Savage) | 2 |  |
| Intercept (Thane, Warrior) | 2 |  |
| Rapid Fire (Hunter) | 2 |  |
| Weaponry: Axes (Berserker, Savage, Shadowblade, Skald, Thane, Viking, Warrior) | 2 | yes |
| Weaponry: Hammers (Berserker, Healer, Savage, Seer, Shaman, Skald, Thane, Viking, Warrior) | 2 | yes |
| Weaponry: Swords (Berserker, Hunter, Midgard Rogue, Savage, Shadowblade, Skald, Thane, Viking, Warrior) | 2 | yes |
| Call of a thousand storms (Thane) | 1 |  |
| Camouflage (Hunter) | 1 |  |
| Climb Walls (Shadowblade) | 1 |  |
| Danger Sense (Shadowblade) | 1 |  |
| Detect Hidden (Shadowblade) | 1 |  |
| Distraction (Shadowblade) | 1 |  |
| Engage (Thane, Warrior) | 1 |  |
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
| DOL.GS.GameGuard | 197 |  |
| DOL.GS.Scripts.Myling | 194 |  |
| DOL.GS.Keeps.GuardFighter | 139 |  |
| DOL.GS.NightSpawn | 71 |  |
| DOL.GS.Scripts.ShiveringPresence | 51 |  |
| DOL.GS.Keeps.GuardArcher | 49 |  |
| DOL.GS.GameHastener | 27 |  |
| DOL.GS.Keeps.GuardCommander | 26 |  |
| DOL.GS.Blacksmith | 23 |  |
| DOL.GS.GameStableMaster | 23 |  |
| DOL.GS.QueenMajorAdd | 22 |  |
| DOL.GS.GameHealer | 18 | yes |
| DOL.GS.Scripts.OFAssistant | 18 |  |
| DOL.GS.Scripts.LiveTeleporter | 15 |  |
| DOL.GS.Enchanter | 14 |  |
| DOL.GS.StripeAdd | 13 |  |
| DOL.GS.ZritZritAdd | 13 |  |
| DOL.GS.Keeps.FrontierHastener | 11 |  |
| DOL.GS.GameVaultKeeper | 9 |  |
| DOL.GS.OonaUndeadSoldier | 9 |  |
| DOL.GS.Keeps.GuardStaticArcher | 8 |  |
| DOL.GS.RainbowSpriteWhite | 8 |  |
| DOL.GS.Trainer.BerserkerTrainer | 8 |  |
| DOL.GS.Trainer.BonedancerTrainer | 8 |  |
| DOL.GS.Trainer.HealerTrainer | 8 |  |
| DOL.GS.Trainer.HunterTrainer | 8 |  |
| DOL.GS.Trainer.RunemasterTrainer | 8 |  |
| DOL.GS.Trainer.ShadowbladeTrainer | 8 |  |
| DOL.GS.Trainer.ShamanTrainer | 8 |  |
| DOL.GS.Trainer.SkaldTrainer | 8 |  |
| DOL.GS.Trainer.SpiritmasterTrainer | 8 |  |
| DOL.GS.Trainer.ThaneTrainer | 8 |  |
| DOL.GS.Trainer.WarriorTrainer | 8 | yes |
| DOL.GS.Keeps.GuardCorpseSummoner | 7 |  |
| DOL.GS.Keeps.GuardLord | 7 |  |
| DOL.GS.RainbowSpriteBlue | 7 |  |
| DOL.GS.RainbowSpriteTan | 7 |  |
| DOL.GS.Trainer.SavageTrainer | 7 |  |
| DOL.GS.RainbowSpriteGreen | 6 |  |
| DOL.GS.Recharger | 5 |  |
| DOL.GS.Trainer.SeerTrainer | 5 |  |
| DOL.GS.Trainer.VikingTrainer | 5 | yes |
| DOL.GS.DPSDummy | 4 |  |
| DOL.GS.Trainer.MysticTrainer | 4 |  |
| DOL.GS.GameBountyMerchant | 3 |  |
| DOL.GS.MidgardSITeleporter | 3 |  |
| DOL.GS.Scripts.OFTeleporter | 3 |  |
| DOL.GS.SnarlsAdd | 3 |  |
| DOL.GS.Trainer.RogueTrainer | 3 |  |
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
| DOL.GS.AccountVaultKeeper | 1 |  |
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
| 31-35 | **none** | **none** | **none** | **none** | **none** | **none** | **none** | - | **none** |
| 36-40 | **none** | **none** | **none** | **none** | **none** | **none** | **none** | - | **none** |
| 41-45 | **none** | **none** | **none** | **none** | **none** | **none** | **none** | - | **none** |
| 46-50 | **none** | **none** | **none** | **none** | **none** | **none** | **none** | - | **none** |

## Rules

| rule | status |
| --- | --- |
| character.class | reference |
| character.creation | documented |
| character.endurance | provisional |
| character.health-regen | provisional |
| character.hit-points | reference |
| character.item-stats | documented |
| character.level-stats | documented |
| character.race-stats | documented |
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
| experience.kill | documented |
| experience.table | documented |
| loot.money | provisional |
| melee.armour | documented |
| melee.base-damage | documented |
| melee.block | documented |
| melee.critical | reference |
| melee.evade | documented |
| melee.location | reference |
| melee.miss | documented |
| melee.miss-bonuses | reference |
| melee.order | reference |
| melee.parry | documented |
| melee.style-cost | reference |
| melee.style-damage | documented |
| melee.swing-speed | documented |
| melee.weapon-skill | reference |
| spec.abilities | documented |
| spec.items | documented |
| spec.migration | documented |
| spec.points | reference |
| spec.training | documented |
| start.town | documented |
| start.unsafe | provisional |
