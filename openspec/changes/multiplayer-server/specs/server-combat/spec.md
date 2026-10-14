## ADDED Requirements

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
Styles, their requirements, endurance costs and effects SHALL come from the imported style and class tables, with parry, evade and block available where the class and weapon allow.

#### Scenario: Opening after a block
- **WHEN** a style requires an opening after a block and the player just blocked
- **THEN** the server allows the style and spends its endurance

#### Scenario: Not enough endurance
- **WHEN** a style costs more endurance than the player has
- **THEN** the server refuses it and says why

### Requirement: Combat is visible to all
Swings, hits, blocks, flinches and deaths SHALL play on every nearby client with the same clips as today.

#### Scenario: Watching a fight
- **WHEN** a player fights a wolf
- **THEN** a second player nearby sees the swings, the wolf's flinch and its death

### Requirement: Death of a player
A dead player SHALL be removed from combat, and SHALL return alive at their bind point after a short delay.

#### Scenario: Killed by wolves
- **WHEN** a player's health reaches zero
- **THEN** NPCs on them go home, and the player stands at their bind point after the delay
