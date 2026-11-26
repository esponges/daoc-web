# server-combat Specification

## Purpose
TBD - created by archiving change multiplayer-server. Update Purpose after archive.

## Requirements

### Requirement: The server resolves combat
The server SHALL decide every swing, hit, block, parry, evade, damage and death, and clients SHALL only request actions and display the results.

#### Scenario: Modified client
- **WHEN** a client claims it hit a target the server says is out of reach
- **THEN** no damage is dealt

#### Scenario: Auto-attack
- **WHEN** a player has a target in reach and arc and attack switched on
- **THEN** the server resolves a swing each time the swing timer elapses and sends the outcome to everyone nearby

### Requirement: Combat rules are implemented independently
Combat rules SHALL be this project's own code, informed by documented game mechanics and by numbers in the imported data, and SHALL NOT copy source from a GPL server emulator.

#### Scenario: Source review
- **WHEN** the combat package is reviewed
- **THEN** it has no code taken from a GPL project and its tests state the formula each case checks

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

#### Scenario: A style starts the fight
- **WHEN** a player with a hostile target uses a style while not attacking
- **THEN** the server turns their attack on and queues the style for the first swing, as the game does; with no target it says so

#### Scenario: The next swing is named
- **WHEN** a style is queued
- **THEN** the player is sent a `prep` event naming it, and the log reads "You prepare to perform <style>."

### Requirement: Combat is visible to all
Swings, hits, blocks, flinches and deaths SHALL play on every nearby client with the same clips as today.

#### Scenario: A style looks like itself
- **WHEN** a player swings with a style
- **THEN** the swing's `s` names the style's row of the install's `styles.csv` (its Icon, or its two-handed animation with a two-handed weapon) and every client plays that style's own clip from `canims.csv`

#### Scenario: A player swings as the weapon held
- **WHEN** a player swings without a style
- **THEN** they play the plain swings and combat stance of the kind of weapon they hold (`wd`: one-handed, two-handed, spear, staff, dual or hand to hand), never the hand-to-hand set of the bare race body when they hold a weapon

#### Scenario: Watching a fight
- **WHEN** a player fights a wolf
- **THEN** a second player nearby sees the swings, the wolf's flinch and its death

### Requirement: Death of a player
A dead player SHALL be removed from combat, and NPCs fighting them SHALL go home. A death to a creature SHALL cost experience at once. The player SHALL then return alive at their bind point when they release or the release timer ends, paying the release penalties of `rules-death`.

#### Scenario: Killed by wolves
- **WHEN** a level 12 player's health reaches zero
- **THEN** NPCs on them go home, they lose experience, and when they release they stand at their bind point, a gravestone marks where they fell and their constitution is lower

#### Scenario: Killed at level 3
- **WHEN** a level 3 player dies
- **THEN** they lose nothing and simply return to their bind point when they release
