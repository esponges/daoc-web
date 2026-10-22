# Spec Delta

## Purpose

Decides where a new Midgard character first appears and binds: at the start town 1.65 gave their race and class, and never beside creatures that kill a level 1 character at once, without editing the imported data.

## ADDED Requirements

### Requirement: New characters start where 1.65 started them
A new character SHALL start at the imported start location for their realm, race and class when a recorded source agrees with it. The source is ZAM's starting-zones table, which, for example, starts Troll Vikings in Galplen in Myrkwood Forest, Norse Vikings in Vasudheim in East Svealand, and Dwarf Vikings in Mularn. The load report SHALL name any Midgard start that the source does not agree with. Status: documented.

#### Scenario: The Troll start
- **WHEN** a Troll Warrior is created
- **THEN** they appear at the imported Troll Viking start, in Galplen in Myrkwood Forest, with the town's trainers and merchants around them

### Requirement: An unsafe start is replaced by the realm's safe start
At load the server SHALL test every imported start location against creatures as they spawn, after any template has replaced their values. A start is unsafe if an aggressive creature of level 6 or more is within 1,500 units of it. Level 6 is five levels above a new character. An unsafe start SHALL be replaced, for new characters, by the realm's safe start: the safe start that the most of the realm's race and class rows use. Status: provisional.

#### Scenario: An unsafe start
- **WHEN** a synthetic start has an aggressive level 26 creature 800 units away
- **THEN** that start is marked unsafe, and new characters of that race and class appear at the realm's safe start instead

#### Scenario: A safe start is kept
- **WHEN** every aggressive creature within 1,500 units of a start is below level 6, or the only stronger ones are passive
- **THEN** new characters of that race and class appear there as the data says

#### Scenario: The Galplen start is safe
- **WHEN** the world loads from a store imported with template values honoured
- **THEN** the Troll Warrior start is not marked unsafe

### Requirement: Starts are reported at load
The server SHALL log, at load, each unsafe start with its race, class, zone and the hardest creature that made it unsafe (name, level and distance), and SHALL log the start chosen in its place. When no start is unsafe it SHALL log one line saying so.

#### Scenario: Load report
- **WHEN** the server starts and a start is unsafe
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

#### Scenario: A level 1 character at an unsafe start
- **WHEN** a level 1 character created before this change logs in at a start the rule marks unsafe
- **THEN** they appear at the safe start, are bound there, and a system line says the start was moved

#### Scenario: A higher character stays
- **WHEN** a level 12 character logs in at that place
- **THEN** they stay where they are and their bind point is unchanged

### Requirement: Dangerous creatures stay visible
Creatures far above a character's level SHALL still be spawned, drawn, named and coloured by con, so the player can see the danger.

#### Scenario: A level 26 creature in view
- **WHEN** a level 1 character looks at a level 26 creature
- **THEN** it is drawn with its name in the purple con colour
