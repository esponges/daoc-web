## ADDED Requirements

### Requirement: Player casts respect line of sight
A player's spell aimed at another combatant SHALL be checked for line of sight when it starts and when it lands, under `rules-line-of-sight`. A spell that fails the check SHALL spend no power and start no recast timer.

#### Scenario: Blocked at the start
- **WHEN** a Shaman starts a damage-over-time spell at a creature they cannot see
- **THEN** the cast does not start, "You can't see your target!" is shown, and the spell stays ready

### Requirement: Resurrection spells raise a dead player
A resurrection spell SHALL target a dead player who has not released, within its range. The dead player SHALL be offered the resurrection and SHALL accept or decline it. When they accept, they SHALL rise where they lie, with the health and power percentages the spell gives. A spell aimed at a living or released player SHALL be refused.

#### Scenario: Raising a group member
- **WHEN** a Healer casts a resurrection with 10% health on a dead group member, who accepts
- **THEN** the member stands up where they died with 10% of their health and power, and is no longer dead

#### Scenario: Declined
- **WHEN** the dead player declines
- **THEN** they stay dead, may still release, and the caster's power is not refunded

#### Scenario: Already released
- **WHEN** the target has already released
- **THEN** the spell is refused with "Your target is not dead" or the game's message for it

### Requirement: Concentration buffs are cast like other buffs
A spell with a concentration cost SHALL be cast through the same checks as any buff, plus the concentration check in `rules-concentration`. The spell bar and SPELLS window SHALL show a held concentration buff as on.

#### Scenario: Buff bot behaviour
- **WHEN** a Shaman casts their strength/constitution concentration buff on each of three group members
- **THEN** each holds the buff with no timer, and the Shaman's concentration in use is three times its cost

### Requirement: Summon and charm spells make pets
Summon spells SHALL create a pet, and charm spells SHALL take a creature as a pet, under `pets`. A minion summon with no commander SHALL be refused.

#### Scenario: A Hunter's pet
- **WHEN** a Hunter casts their Beastcraft summon
- **THEN** their pet appears and the pet window opens

### Requirement: Spells aimed at the caster's pet can be cast
A spell whose target is the pet SHALL fall on the caster's own pet (a Bonedancer's commander; never a minion), and be refused with "You must have a pet to cast that." when there is none. Every lasting effect a creature can hold (strength and constitution, dexterity and quickness, combat speed, damage add, damage shield, ablative armour, mesmerize duration, an offensive proc) SHALL work on a pet in its own fights, and a heal SHALL raise it.

#### Scenario: Pet spell without a pet
- **WHEN** a Spiritmaster with no pet casts a buff aimed at their pet
- **THEN** the cast is refused and no power is spent

#### Scenario: A pet proc
- **WHEN** a pet holding an offensive proc strikes a creature and the proc's chance comes up
- **THEN** the proc's spell is cast on the creature by the pet

### Requirement: Debuffs, cures and the remaining Midgard buffs can be cast
Every spell type that the twelve Midgard classes' lines use SHALL be castable by a player, under the effects in `rules-spell-effects`: the remaining buffs, the debuffs and control debuffs, the cures and spread heal, life transfer, and damage with a debuff or slow. Any type left over SHALL be counted in the census with a reason.

#### Scenario: A Bonedancer's strength debuff
- **WHEN** a Bonedancer casts a strength debuff on a creature that does not resist
- **THEN** the creature's strength falls by the spell's value for its duration, and its melee damage falls with it

#### Scenario: Cure Poison
- **WHEN** a Healer casts Cure Poison on a group member carrying a poison
- **THEN** the poison effect ends on them
