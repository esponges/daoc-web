## ADDED Requirements

### Requirement: The compass agrees with the map
The compass SHALL show north in the direction the zone's own map calls north, checked against at least one landmark or zone exit whose compass direction is known.

#### Scenario: Facing a known exit
- **WHEN** the camera faces an exit whose direction from the character is known from the zone data
- **THEN** the compass shows that direction in the middle

### Requirement: The clock does not invent a time
The clock SHALL show a game time only if the install defines a day length, and otherwise SHALL show the player's own time and say so in its tooltip.

#### Scenario: No day length in the install
- **WHEN** no day length can be established from the install
- **THEN** the clock shows local time and its tooltip says so
