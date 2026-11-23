## ADDED Requirements

### Requirement: Hibernia classes are playable
Each of Hibernia's thirteen 1.65 classes SHALL be creatable with the races the start table allows. Each SHALL be trainable, and SHALL use its lines, styles, spells and abilities from the community data. Gaps SHALL be listed with reasons in the census.

#### Scenario: A Blademaster
- **WHEN** a player makes a Celt Blademaster and trains Celtic Dual
- **THEN** they swing with both hands under the dual-wield rules

### Requirement: Animist turrets
An Animist SHALL place turrets at a ground target within range. Turrets SHALL not move, SHALL attack or buff by their spell, and SHALL count against the Animist's turret limit. Turrets SHALL leave when the Animist moves out of range, dies or dismisses them.

#### Scenario: Placing a turret
- **WHEN** an Animist with a ground target 800 units away summons a turret
- **THEN** the turret appears at the ground target and attacks the Animist's target

#### Scenario: Limit
- **WHEN** an Animist at the turret limit summons another
- **THEN** it is refused, naming the limit

### Requirement: Start towns
A new Hibernia character SHALL start in the town that the start table gives for their race and class.

#### Scenario: Start
- **WHEN** a player makes a Lurikeen Eldritch
- **THEN** they enter the world at the start town the table gives
