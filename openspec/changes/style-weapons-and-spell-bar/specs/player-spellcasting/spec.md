## ADDED Requirements

### Requirement: The spell bar shows what stops a cast
A spell bar slot SHALL be shaded red while the character cannot pay for its spell: power for a power spell, hit points for a spell paid in hit points (the cost may not take them to 0). When a spell with a recast lands, the server SHALL send `recast` with the spell and its seconds, and the slot SHALL be darkened with the seconds left until the recast has passed.

#### Scenario: A Savage's spell with health to spare
- **WHEN** a Savage with 300 of 400 hit points has Nails of Kelgor (12 hit points) on the bar
- **THEN** the slot is not shaded red

#### Scenario: A recast runs out
- **WHEN** the Savage casts Nails of Kelgor, which has a 5 second recast
- **THEN** the slot is darkened and counts down from 5, and after 5 seconds it is clear again
