## ADDED Requirements

### Requirement: Setting a ground target
A player SHALL be able to set a ground target on the terrain or a floor they can see, within 2000 units of where they stand, and the server SHALL keep it until it is replaced or cleared, or the player changes zone.

#### Scenario: Set in range
- **WHEN** a player sets a ground target 1500 units away on visible ground
- **THEN** the server keeps it and the client shows a marker there

#### Scenario: Out of range
- **WHEN** a player tries to set a ground target 2500 units away
- **THEN** it is refused with "Your ground target is too far away!" and the old target stays

#### Scenario: Zone change clears it
- **WHEN** a player with a ground target changes zone
- **THEN** the ground target is cleared

### Requirement: Use by abilities and spells
An ability or spell that needs a ground target SHALL be refused when none is set or it is out of the ability's range. It SHALL otherwise take effect at the ground target.

#### Scenario: No ground target
- **WHEN** a Shadowblade uses Distraction with no ground target
- **THEN** it is refused with "You must have a ground target to use this!"

#### Scenario: With a ground target
- **WHEN** a Shadowblade uses Distraction on a ground target in range
- **THEN** the effect happens at that point and the reuse timer starts
