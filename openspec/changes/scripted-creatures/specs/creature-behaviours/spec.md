## ADDED Requirements

### Requirement: Time-of-day presence
A creature kind marked for night or day SHALL be present only in that part of the game clock. It SHALL leave the world when its time ends unless it is in combat, and return at its point when its time comes again. The parity census SHALL count these kinds (NightSpawn, DaySpawn) as implemented.

#### Scenario: NightSpawn at dawn
- **WHEN** the game clock passes dawn and a NightSpawn creature is not fighting
- **THEN** it leaves the world, and returns at dusk

### Requirement: Sourced or marked
A creature kind SHALL get behaviour beyond its template only where a public source says what it does. A kind with no source SHALL spawn as a plain creature with its database template, and the parity census SHALL list it with its reason: behaviour unsourced, owned by a later change, or not classic.

#### Scenario: Census
- **WHEN** `cmd/parity` runs
- **THEN** every Midgard creature kind that is not implemented has a reason: owned by a named later change, an unsettled item service, not classic, or behaviour unsourced

#### Scenario: No guessed behaviour
- **WHEN** a named creature has no public source for its behaviour
- **THEN** it behaves as its template says and nothing more
