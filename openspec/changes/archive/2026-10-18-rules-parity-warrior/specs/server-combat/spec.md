# Spec Delta

## MODIFIED Requirements

### Requirement: Styles and defences come from data
Styles, their requirements, endurance costs and effects SHALL come from the imported style and class tables. Parry, evade and block SHALL be available where the class's specs and abilities and the held equipment allow. The numbers they produce SHALL be those of `rules-melee`.

#### Scenario: Opening after a block
- **WHEN** a style requires an opening after a block and the player just blocked
- **THEN** the server allows the style and spends its endurance by the weapon's speed

#### Scenario: Not enough endurance
- **WHEN** a style costs more endurance than the player has
- **THEN** the server refuses it and says why

#### Scenario: A style that is parried
- **WHEN** a queued style's swing is parried
- **THEN** the style does not fire and half its endurance is spent

### Requirement: Death of a player
A dead player SHALL be removed from combat, and NPCs fighting them SHALL go home. A death to a creature SHALL cost experience at once. The player SHALL then return alive at their bind point when they release or the release timer ends, paying the release penalties of `rules-death`.

#### Scenario: Killed by wolves
- **WHEN** a level 12 player's health reaches zero
- **THEN** NPCs on them go home, they lose experience, and when they release they stand at their bind point, a gravestone marks where they fell and their constitution is lower

#### Scenario: Killed at level 3
- **WHEN** a level 3 player dies
- **THEN** they lose nothing and simply return to their bind point when they release
