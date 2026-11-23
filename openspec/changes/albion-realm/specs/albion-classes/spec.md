## ADDED Requirements

### Requirement: Albion classes are playable
Each of Albion's fourteen 1.65 classes SHALL be creatable with the races the start table allows. Each SHALL be trainable at its trainer, and SHALL use its specialization lines, styles, spells and abilities from the community data. Gaps SHALL be listed with reasons in the census.

#### Scenario: A Cleric
- **WHEN** a player makes a Briton Cleric and trains Rejuvenation
- **THEN** they can cast its heals

#### Scenario: Census
- **WHEN** `cmd/parity` runs
- **THEN** the Albion section lists class spell types and abilities with implemented counts and reasons

### Requirement: Necromancer shade form
A Necromancer with a summoned pet SHALL become a shade: the player cannot be hit directly, damage aimed at them goes to the pet, and the player casts through the pet. When the pet dies, the shade form ends and the Necromancer takes the shade-ending penalty in the rule book.

#### Scenario: Attacking the shade
- **WHEN** a creature attacks a Necromancer in shade form
- **THEN** the pet takes the damage

### Requirement: Start towns
A new Albion character SHALL start in the town that the start table gives for their race and class.

#### Scenario: Start
- **WHEN** a player makes an Avalonian Wizard
- **THEN** they enter the world at the start town the table gives
