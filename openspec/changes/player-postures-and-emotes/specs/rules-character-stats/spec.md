## MODIFIED Requirements

### Requirement: Endurance is a pool of 100
A character's maximum endurance SHALL be 100. Endurance SHALL return 6 points a second out of combat and 2 a second in combat; a seated character out of combat SHALL regain twice the out-of-combat amount (`rules-resting`). Status: provisional for the regeneration rates.

#### Scenario: After a style chain
- **WHEN** a Warrior with 40 endurance stops fighting
- **THEN** they are back to 100 within ten seconds

#### Scenario: Seated
- **WHEN** a Warrior with 40 endurance sits out of combat
- **THEN** they are back to 100 within five seconds

### Requirement: Health returns slowly out of combat
Out of combat, a living character SHALL regain 2.5 + 0.25 × level hit points every 3 seconds, rounded down and at least 1, and twice that while seated (`rules-resting`). In combat they SHALL regain none. Status: provisional for the tick and the in-combat rule, reference for the amount.

#### Scenario: Level 20 resting
- **WHEN** a level 20 Warrior below full health is out of combat for 30 seconds
- **THEN** they regain 70 hit points, or 140 if seated
