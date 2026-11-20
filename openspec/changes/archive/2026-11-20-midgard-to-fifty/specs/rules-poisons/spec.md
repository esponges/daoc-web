## Purpose

Defines how a Shadowblade's Envenom poisons work: buying a poison, applying it to a weapon, its charges, and how it fires on a hit.

## ADDED Requirements

### Requirement: A poison is applied to a weapon
A player with the Envenom spec SHALL be able to apply a poison item to a weapon they wear in the right or left hand. The poison item SHALL be used up. Applying it SHALL be refused when the poison's level is above what the player's Envenom spec allows. A newly applied poison SHALL replace the one already on the weapon. Status: reference.

#### Scenario: Applying a poison
- **WHEN** a Shadowblade with Envenom 15 applies a level 15 poison to their main-hand axe
- **THEN** the poison item leaves their pack, and the weapon shows the poison and its charges

#### Scenario: Not enough Envenom
- **WHEN** a Shadowblade with Envenom 5 applies a level 20 poison
- **THEN** it is refused with a message saying their skill is too low

### Requirement: A poison fires on a hit and uses a charge
A poisoned weapon SHALL cast its poison's spell on the target when a swing with it hits, at the chance the rule gives, and each firing SHALL use one charge. With no charges left, the poison SHALL be gone. The poison's spell SHALL apply through the same spell engine as procs, with resists. Status: set by the findings.

#### Scenario: Damage-over-time poison
- **WHEN** a poisoned axe hits a creature and the poison fires
- **THEN** the creature carries the poison's damage over time, and the weapon has one charge fewer

#### Scenario: Last charge
- **WHEN** the last charge is used
- **THEN** the weapon no longer shows a poison, and later hits cast nothing

### Requirement: Poisons can be bought in Midgard
Midgard's poison merchants SHALL sell the poisons their stock lists name, so a Shadowblade can buy each poison kind their Envenom spec opens.

#### Scenario: A poison merchant
- **WHEN** a Shadowblade opens a merchant whose list holds poisons
- **THEN** the poisons are offered and can be bought like any other item
