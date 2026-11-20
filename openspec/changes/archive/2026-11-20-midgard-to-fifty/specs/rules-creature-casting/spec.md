## ADDED Requirements

### Requirement: Creatures need line of sight to cast
A creature SHALL cast a hostile spell only at a target it can see under `rules-line-of-sight`. When it cannot, it SHALL close in, and it SHALL NOT spend the spell's recast.

#### Scenario: Through a wall
- **WHEN** a casting creature's target is behind a dungeon wall within spell range
- **THEN** no spell is cast at the target, and the creature moves toward them

### Requirement: Creatures cast their buffs and debuffs
A creature whose template lists buffs (armour factor, strength, strength/constitution) SHALL cast them on itself when it enters combat and they are not on. A creature whose template lists debuffs (strength, dexterity) SHALL cast them at its target when they are ready. All of these SHALL follow `rules-spell-effects`.

#### Scenario: Self armour buff
- **WHEN** a creature carrying an armour factor buff spell is pulled
- **THEN** it casts the buff on itself, and its armour factor rises by the spell's value

#### Scenario: Debuff on a player
- **WHEN** a creature with a dexterity debuff fights a player and the spell is ready
- **THEN** it casts the debuff at the player, who may resist it
