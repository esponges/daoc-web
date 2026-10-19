# Spec Delta

## Purpose

Decides where a new Midgard character first appears and binds, so that no one starts beside creatures that kill a level 1 character at once, without editing the imported data.

## ADDED Requirements

### Requirement: An unsafe start is replaced by the realm's safe start
At load the server SHALL test every imported start location. A start is unsafe if an aggressive creature of level 6 or more is within 1,500 units of it. Level 6 is five levels above a new character. An unsafe start SHALL be replaced, for new characters, by the realm's safe start, chosen as described in the next requirement. Status: provisional.

#### Scenario: The Troll start
- **WHEN** the world loads with the imported Troll Warrior start in Myrkwood Forest, near level 26 aggressive creatures
- **THEN** that start is marked unsafe and a new Troll Warrior appears at Midgard's safe start instead

#### Scenario: A safe start is kept
- **WHEN** every aggressive creature within 1,500 units of a start is below level 6
- **THEN** new characters of that race and class appear there as the data says

### Requirement: The safe start has a source or a rule
A realm's safe start SHALL be the start town that 1.65 used for that race and class, when a recorded source names it. Otherwise it SHALL be the start, among the realm's safe starts, that the most race and class rows use. Status: provisional, until a source is recorded.

#### Scenario: No source yet
- **WHEN** no 1.65 start-town source has been recorded
- **THEN** Midgard's safe start is the safe start most Midgard rows use, and the load report names it

### Requirement: Starts are reported at load
The server SHALL log, at load, each unsafe start with its race, class, zone and the hardest creature that made it unsafe (name, level and distance), and SHALL log the start chosen in its place.

#### Scenario: Load report
- **WHEN** the server starts
- **THEN** the log has one line per replaced start naming the creature, its level and distance, and the replacement

### Requirement: The imported data is not changed
The rule SHALL change only where the server places new characters. The world store's start rows, spawns, levels and aggro values SHALL stay as imported.

#### Scenario: Store untouched
- **WHEN** the rule replaces a start
- **THEN** the `startup` and `spawn` rows in `world.db` are byte-for-byte what the importer wrote

### Requirement: Characters keep a bind point
Each character SHALL have a stored bind point. A new character's bind point SHALL be their start. `/bind` within the radius of an imported bind point of their realm SHALL set it there, and elsewhere SHALL be refused. A dead character SHALL return to their bind point. A character with no stored bind point SHALL be bound at the bind point nearest to where they stand when they first log in after this change. Status: reference.

#### Scenario: Binding in a town
- **WHEN** a player stands at the Audliten bindstone and types `/bind`
- **THEN** they are told they are bound there, and their next death returns them to Audliten

#### Scenario: Away from a bindstone
- **WHEN** a player types `/bind` in open country
- **THEN** they are told there is nothing to bind to here, and the bind point is unchanged

### Requirement: Low-level characters at an unsafe start are moved once
A character of level 5 or less who logs in within 1,500 units of an unsafe start, or whose bind point is one, SHALL be moved and bound to the replacement start, and SHALL be told so once.

#### Scenario: An existing level 1 Troll
- **WHEN** a level 1 Troll created before this change logs in at the Myrkwood start
- **THEN** they appear at the safe start, are bound there, and a system line says the start was moved

#### Scenario: A higher character stays
- **WHEN** a level 12 character logs in at that place
- **THEN** they stay where they are and their bind point is unchanged

### Requirement: Dangerous creatures stay visible
Creatures far above a character's level SHALL still be spawned, drawn, named and coloured by con, so the player can see the danger.

#### Scenario: A level 26 creature in view
- **WHEN** a level 1 character looks at a level 26 creature
- **THEN** it is drawn with its name in the purple con colour
