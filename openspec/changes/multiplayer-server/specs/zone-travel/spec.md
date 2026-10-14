## ADDED Requirements

### Requirement: Scenery blocks movement
Characters and NPCs SHALL not walk through trees, buildings, fences, or the walls and floors of cities and dungeons, using the collision hulls in the models.

#### Scenario: Walking into a building
- **WHEN** a player walks into the wall of a building in Mularn
- **THEN** the character slides along it or stops, and the server and client agree on where it is

#### Scenario: Wall in a dungeon
- **WHEN** a player in a dungeon walks at a wall
- **THEN** they do not pass through it or drop to a lower level

#### Scenario: Model with no hull
- **WHEN** a model has no collision hull
- **THEN** it does not block movement

### Requirement: Exits move a character between zones
Walking into a zone exit SHALL move the character to the exit's target zone and position, and SHALL be refused if the target zone is not available.

#### Scenario: Enter a dungeon
- **WHEN** a player walks into Spindelhalla's entrance exit
- **THEN** they appear at its entrance inside zone 125 and other players there see them

#### Scenario: Target not converted
- **WHEN** an exit leads to a zone with no converted data
- **THEN** the character stays where it is and is told why

### Requirement: Zones run independently
The server SHALL simulate only the zones that have players, and SHALL save and unload an empty zone's state.

#### Scenario: Last player leaves
- **WHEN** the last player leaves a zone
- **THEN** the zone stops ticking and its NPCs return to their initial state when next loaded

### Requirement: Edges of outdoor zones lead to their neighbours
Walking off the edge of an outdoor zone SHALL enter the adjoining zone of the same region, if it is converted, and set the character down just inside it.

#### Scenario: West from Myrkwood Forest
- **WHEN** a character walks off Myrkwood Forest's west edge where Skona Ravine adjoins it
- **THEN** they are saved in Skona Ravine near its east edge and their client loads it

#### Scenario: Nothing beyond the edge
- **WHEN** no converted zone adjoins that part of the edge
- **THEN** the character stays where it is

### Requirement: An exit is walked into
An exit SHALL take a character through only when they walk into it, and a character who arrives standing in one SHALL have to step out of it first.

#### Scenario: Arriving in a dungeon
- **WHEN** a character comes through an exit and arrives within the twin exit's trigger distance
- **THEN** they are not sent back until they have stepped out of it and in again

### Requirement: Indoors, ledges and voids are solid
Inside a city or dungeon a move that would leave the floor, onto a drop of more than 110 units or into a void, SHALL be refused or slid along, as a wall is.

#### Scenario: Running at a ledge
- **WHEN** a character runs in every direction from the entrance of Spindelhalla for ten seconds
- **THEN** their height never changes by more than a step and a half in one move, and they stay inside the dungeon
