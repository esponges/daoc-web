## ADDED Requirements

### Requirement: Speed enhancement raises speed
A speed-enhancement spell SHALL multiply its holder's movement speed by its `Value` read as a percent (144 runs at 1.44 times). The strongest active one SHALL count, and it SHALL multiply with any slow. It SHALL not take hold on a target that is in combat, and SHALL end when its holder enters combat. The client's predicted movement SHALL agree with the server at the enhanced speed. Status: reference.

#### Scenario: Simple Song of Travel
- **WHEN** a 144 Song of Travel is on a running character with no slow
- **THEN** they run at 1.44 times their normal speed, and the shared movement vector for that factor passes in Go and JS

#### Scenario: The strongest counts
- **WHEN** a player holds a 144 and a 159 speed enhancement
- **THEN** they run at 1.59 times their normal speed

#### Scenario: With a slow
- **WHEN** a character with a 144 song is hit by a 40% slow
- **THEN** their speed is 1.44 × 0.6 of normal

#### Scenario: Combat
- **WHEN** a character with a song on swings at a creature
- **THEN** the song's effect ends on them, and a new pulse does not take hold until they are out of combat

### Requirement: A health regeneration buff adds to health ticks
A health-regeneration spell SHALL add its `Value` hit points to each health tick while it holds, and the strongest one SHALL count. Under the rule that no health returns in combat, the buff adds nothing in combat. Status: provisional.

#### Scenario: Song of Rest out of combat
- **WHEN** a level 20 character out of combat holds a 10-point Song of Rest
- **THEN** each health tick returns 10 more hit points than it would without the song

#### Scenario: In combat
- **WHEN** the same character is in combat with the song on
- **THEN** no tick returns hit points

### Requirement: Resist buffs raise a resist
A resist buff SHALL add its `Value` in percent to its holder's resist against its damage type: body, spirit, energy, heat, cold or matter, whatever the caster's stat or spec. The strongest buff for a damage type SHALL count, and resist buffs SHALL add at most 24% to one resist. The total SHALL still be capped by the resist cap of 70%. Status: reference.

#### Scenario: A heat chant against a fire bolt
- **WHEN** a player with 5% heat resist from their race holds a 10-point Heat Diminishing Chant and is hit by a heat spell
- **THEN** the spell meets 15% heat resist

#### Scenario: Another damage type is untouched
- **WHEN** the same player is hit by a cold spell
- **THEN** the heat chant adds nothing to the resist it meets

#### Scenario: The buff cap
- **WHEN** a player holds a 30-point body resist buff
- **THEN** it adds 24% to their body resist

### Requirement: Area damage does not fall off
Area damage SHALL not fall off with distance from the centre: each creature inside the radius takes its own full roll. The centre SHALL be the caster when the spell's range is 0, and the target otherwise. Status: provisional (the community servers differ and no 1.65 account was found).

#### Scenario: Centre and edge
- **WHEN** an area spell hits one creature at its centre and one at its edge
- **THEN** each takes a damage roll from the same range, resisted by its own resist
