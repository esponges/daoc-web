# rules-creature-combat Specification

## Purpose
Defines the fighting numbers of ordinary creatures (hit points, weapon skill, damage, armour, swing speed, defences and resists) from their level, stats, template and race, as the classic reference server derives them.

## Requirements

### Requirement: Creature stats follow their level
An ordinary creature's strength, constitution, dexterity and quickness SHALL each be 30 + (level − 1). This SHALL hold whatever the spawn row says. Status: reference (the reference server derives these from level and ignores the stored values).

#### Scenario: Level 10 creature
- **WHEN** a level 10 wolf spawns
- **THEN** its strength, constitution, dexterity and quickness are 39

### Requirement: Creature hit points
A creature's maximum hit points SHALL be trunc((constitution − 25) × 1.8 + 11 × (level + 1) + 0.6 × level²). Status: reference.

#### Scenario: Hit points by level
- **WHEN** creatures of level 1, 10 and 50 spawn
- **THEN** they have 31, 206 and 2,158 hit points

### Requirement: Creature blows
A creature's base damage SHALL be (1 + level / 2.5 + level² / 120) × weapon speed. Weapon speed is 3.0 s, or 4.0 s when its equipment puts a weapon in the two-handed slot. The damage cap SHALL be 3 × base. Its weapon skill SHALL be level × 440 × 0.005 × (1 + strength × 0.005), with variance uniform from 0.9 to 1.1. The armour term and the final damage SHALL use the same rule as a player's hit. Status: reference.

#### Scenario: A level 10 creature hits a studded torso
- **WHEN** a level 10 creature hits a level 10 Warrior's studded torso (armour factor 24 with the level share, absorb 19%) at a variance roll of 0.5
- **THEN** the hit does 15

#### Scenario: Bare skin hits the cap
- **WHEN** the same blow lands on a location with no armour (armour factor 4, no absorption)
- **THEN** it is capped at 52

### Requirement: Creature armour
A creature's armour factor SHALL be trunc((1 + level / 200) × level × 12.5) / 5 on every location. Its absorption SHALL be level × 0.54%. Status: reference.

#### Scenario: Level 50 creature
- **WHEN** a player hits a level 50 creature
- **THEN** the creature's armour factor is 156.2 and its absorption 27%

### Requirement: Creature swing speed
A creature SHALL swing every weapon speed × (1 − (quickness − 50) × 0.002) seconds, and never faster than every 0.5 s. Status: reference.

#### Scenario: Level 10 one-handed creature
- **WHEN** a level 10 creature with quickness 39 fights
- **THEN** it swings every 3.066 seconds

### Requirement: Creature resists come from its race
Damage to a creature SHALL be reduced by its race's resist to the weapon's damage type, crush, slash or thrust, in percent. Race comes from the spawn's race and the imported race table. A creature with no race row SHALL have no resists. Status: documented (the data).

#### Scenario: A thrust weakness
- **WHEN** a sword's thrust hit lands on a creature whose race resists thrust by −15%
- **THEN** it does 15% more damage, up to the cap

### Requirement: Creature defences come from the template
A creature SHALL evade, parry and block blows from within 60° of its facing at its template's evade, parry and block percents. A creature with no template SHALL not defend. Status: documented (the data).

#### Scenario: A parrying bandit
- **WHEN** a player attacks the front of a creature whose template parries 10%
- **THEN** 10% of swings that pass the evade roll are parried
