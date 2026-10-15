## ADDED Requirements

### Requirement: A new character does not start among creatures that kill it at once
Creatures that would defeat a character in a few blows SHALL NOT aggro it, or be within aggro range of the place it appears, in a starting zone.

#### Scenario: Standing at the spawn
- **WHEN** a level 1 character stands where it first appears in Myrkwood Forest for five minutes
- **THEN** no creature of more than a few levels above it attacks it

#### Scenario: Walking a short way
- **WHEN** the character walks fifty metres from the spawn
- **THEN** it is not aggroed by creatures far above its level

### Requirement: The rule is the server's, not an edit to the data
The imported creature data SHALL NOT be altered to achieve this; the server's use of it (aggro levels, ranges, which spawns are placed) SHALL be what changes, and the change SHALL be reported when the world loads.

#### Scenario: A skipped spawn
- **WHEN** a spawn is left out or made passive by the rule
- **THEN** the load report counts it with the reason

### Requirement: Dangerous creatures stay visible
Creatures far above a character's level SHALL still be drawn, named and coloured by how hard they are, so that the player can see the danger.

#### Scenario: A level 26 creature in view
- **WHEN** such a creature is in view of a level 1 character
- **THEN** its name is drawn in the red of the hardest con colour
