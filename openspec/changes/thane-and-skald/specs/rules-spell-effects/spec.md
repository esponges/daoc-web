## ADDED Requirements

### Requirement: Speed enhancement raises speed
A speed-enhancement spell SHALL multiply its holder's movement speed by its `Value` read as a percent (144 runs at 1.44 times). The strongest active one SHALL count. When a slow also holds, the two SHALL combine as the sources recorded in the findings say. Whether combat ends the effect SHALL follow the sources, and the status SHALL say which. The client's predicted movement SHALL agree with the server at the enhanced speed.

#### Scenario: Simple Song of Travel
- **WHEN** a Skald's 144 Song of Travel is on a running character with no slow
- **THEN** they run at 1.44 times their normal speed, and the shared movement vector for that factor passes in Go and JS

#### Scenario: The strongest counts
- **WHEN** a player holds a 144 and a 159 speed enhancement
- **THEN** they run at 1.59 times their normal speed

### Requirement: A health regeneration buff adds to health ticks
A health-regeneration spell SHALL add its `Value` to the hit points each health tick returns while it holds. Whether it also works in combat SHALL follow the sources in the findings. The strongest one SHALL count. Status: as the findings support; provisional otherwise.

#### Scenario: Song of Rest out of combat
- **WHEN** a level 20 character out of combat holds a 10-point Song of Rest
- **THEN** each health tick returns 10 more hit points than it would without the song

### Requirement: Resist buffs raise a resist
A resist buff SHALL add its `Value` in percent to its holder's resist against its damage type: body, spirit, energy, heat, cold or matter. The strongest buff for a damage type SHALL count. The total SHALL still be capped by the resist cap.

#### Scenario: A heat chant against a fire bolt
- **WHEN** a player with 5% heat resist from their race holds a 10-point Heat Diminishing Chant and is hit by a heat spell
- **THEN** the spell meets 15% heat resist

#### Scenario: Another damage type is untouched
- **WHEN** the same player is hit by a cold spell
- **THEN** the heat chant adds nothing to the resist it meets

### Requirement: Area damage at the edge
Area damage SHALL either fall off with distance from the centre or not, as the sources recorded in the findings say. A creature at the centre SHALL take the full roll. The rule SHALL be registered with its status.

#### Scenario: Centre and edge
- **WHEN** an area spell hits one creature at its centre and one at its edge
- **THEN** the centre creature's damage is the full roll, and the edge creature's damage is what the recorded rule gives for that distance
