## ADDED Requirements

### Requirement: Stick
`/stick` (and the Command window's Stick) SHALL keep the player within melee range of their target, player or creature, facing it, until they press a movement key, the target dies or goes beyond range, or they give the command again.

#### Scenario: Sticking to a fleeing creature
- **WHEN** a player sticks to a creature that runs away
- **THEN** the player follows it, stays in melee range and keeps facing it

#### Scenario: Breaking off
- **WHEN** the player presses a movement key while sticking
- **THEN** sticking ends with a message
