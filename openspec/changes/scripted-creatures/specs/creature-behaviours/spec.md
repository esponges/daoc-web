## ADDED Requirements

### Requirement: Time-of-day presence
A creature kind marked for night or day SHALL be present only in that part of the game clock. It SHALL despawn when its time ends, unless it is in combat, and respawn at its point when its time comes again.

#### Scenario: NightSpawn at dawn
- **WHEN** the game clock passes dawn and a NightSpawn creature is not fighting
- **THEN** it leaves the world, and returns at dusk

### Requirement: Adds
A creature with adds SHALL spawn its add creatures when its health first falls below each threshold in its table. The adds SHALL attack whoever is on the boss's hate list, and SHALL leave when the boss dies or resets.

#### Scenario: Boss at half health
- **WHEN** a boss whose table has an add wave at 50 percent falls below half health
- **THEN** the wave spawns around it and attacks its attackers, once per fight

### Requirement: Fleeing and calling for help
A creature with this behaviour SHALL flee below its health threshold and call same-faction creatures within its call range into the fight.

#### Scenario: Myling calls for help
- **WHEN** a Myling is attacked near other Mylings
- **THEN** the nearby Mylings join the fight

### Requirement: Scripted abilities
A creature with a scripted ability SHALL use it on its timer while in combat. The ability SHALL use the existing spell or style effects.

#### Scenario: Timed area spell
- **WHEN** a named creature with a 20-second area cold spell fights a group
- **THEN** the spell lands on the group every 20 seconds while the fight lasts

### Requirement: Sourced or marked
Every scripted kind SHALL record the source for its behaviour. A kind with no source SHALL spawn with its database template and no extra behaviour, and the census SHALL list it as "behaviour unsourced".

#### Scenario: Census
- **WHEN** `cmd/parity` runs
- **THEN** every Midgard creature kind is implemented, not classic, behaviour unsourced, or owned by a named later change
