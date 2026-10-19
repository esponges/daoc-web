# rules-specialization Specification

## Purpose
Defines how a character earns specialisation points, spends them at a trainer, and what their spec levels and level-granted abilities let them do, as classic 1.65 does.

## Requirements

### Requirement: Levels grant specialisation points
On reaching each level from 2 to 5, a character SHALL gain points equal to that level. From level 6 they SHALL gain ⌊multiplier × level / 10⌋, using the class's multiplier (20 for the Warrior, so 2 × level). From level 40, reaching the midpoint of a level's experience (the half-level) SHALL grant ⌊multiplier × level / 20⌋ more. Status: reference (it agrees with the published 2.0 points per level for pure tanks).

#### Scenario: Level 6 Warrior
- **WHEN** a Warrior reaches level 6
- **THEN** they gain 12 points

#### Scenario: Half-level
- **WHEN** a level 44 Warrior passes the midpoint of level 44's experience
- **THEN** they gain 44 points and are shown as level 44.5

#### Scenario: Total at 50
- **WHEN** a Warrior has reached level 50
- **THEN** they have earned 2,979 points in all, counting 14 from levels 2 to 5

### Requirement: Training a spec costs the level being bought
Every spec a class has SHALL start at level 1. Raising a spec from `n` to `n + 1` SHALL cost `n + 1` unspent points. A spec SHALL NOT be raised above the character's level. Points not spent SHALL be kept. Status: documented.

#### Scenario: Sword to 10
- **WHEN** a level 10 Warrior raises sword from 1 to 10
- **THEN** it costs 2 + 3 + ... + 10 = 54 points

#### Scenario: Above level
- **WHEN** a level 10 Warrior with sword at 10 asks for 11
- **THEN** it is refused with the reason, and no points are spent

#### Scenario: Not enough points
- **WHEN** the next level of a spec costs more than the unspent points
- **THEN** it is refused, and nothing changes

### Requirement: Training happens at a trainer of the class
A character SHALL train only while within interact range of a living trainer NPC for their class or their class's base class (the Viking, for a Warrior). Interacting with a trainer SHALL send the character's specs, levels, costs of the next level and unspent points. A training request SHALL raise one spec by one or more levels. It SHALL be atomic with saving the character. Status: documented.

#### Scenario: At the Warrior trainer
- **WHEN** a Warrior right-clicks a Warrior trainer and trains hammer twice
- **THEN** hammer rises by two levels, the points are spent, the character is saved, and their style bar and sheet are resent

#### Scenario: Away from a trainer
- **WHEN** a training request arrives while the player is not near their class's trainer
- **THEN** it is refused

#### Scenario: The base class's trainer
- **WHEN** a Warrior interacts with a Viking trainer
- **THEN** it offers their specs and trains them as a Warrior trainer would

#### Scenario: Another class's trainer
- **WHEN** a Warrior interacts with a Thane trainer
- **THEN** they are told that trainer cannot teach them, and no specs are offered

### Requirement: Items add to specs in combat
Skill bonuses on worn items, such as +2 sword, SHALL add to that spec wherever combat uses it: variance, the two-handed modifier, parry and block. The total item bonus to one spec SHALL be capped at 1 + ⌊level / 5⌋. Item bonuses SHALL NOT count for training costs, for the level cap on training, or for unlocking styles. Status: documented.

#### Scenario: A +3 sword at level 10
- **WHEN** a level 10 Warrior with sword 8 wears items giving +5 sword
- **THEN** their sword counts as 11 in combat, because only +3 counts, and their style bar still follows sword 8

### Requirement: Styles unlock by spec level
A style SHALL be available when the character's level in the style's spec is at least the style's spec level. The character must also hold what the style needs: the weapon type, or a shield for a Shields style. Character level alone SHALL NOT unlock styles. Status: documented.

#### Scenario: Untrained hammer
- **WHEN** a level 12 Warrior with hammer at 1 holds a hammer
- **THEN** only the hammer styles of spec level 1 or below are on their bar

#### Scenario: Training opens a style
- **WHEN** they raise hammer to 4 and a hammer style needs 4
- **THEN** it appears on the bar at once

### Requirement: Abilities come from the class's careers and specs
A character's abilities, with their levels, SHALL be the ones the imported class and spec tables grant at or below the character's level and spec levels. These include weapon families, armour level, shield size and Evade. When a table grants the same ability more than once, the highest level granted SHALL apply. Status: documented (the data).

#### Scenario: Chain at level 10
- **WHEN** a Warrior reaches level 10
- **THEN** their armour ability rises from studded to chain, and they may now wear chain

#### Scenario: Evade at level 15
- **WHEN** a Warrior reaches level 15
- **THEN** they gain Evade at level 1 and can evade blows from the front

#### Scenario: An ability not yet implemented
- **WHEN** a Warrior reaches level 7 and the tables grant Protect
- **THEN** the ability is listed on their sheet as known but inactive, and the census counts it as unimplemented

### Requirement: Spec state is kept and migrated
Spec levels and unspent points SHALL be saved with the character. A character that predates specs SHALL start with every spec at 1, and with unspent points equal to all the points their level has earned.

#### Scenario: An existing level 20 Warrior
- **WHEN** a level 20 Warrior saved before this change logs in
- **THEN** their specs are all 1 and they have 404 unspent points
