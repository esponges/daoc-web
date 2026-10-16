## ADDED Requirements

### Requirement: Base model follows race and gender
The client SHALL draw a player, and an NPC that has a race and gender, with the base model and textures for that race and gender when the install gives one. Creatures keep their own model. Where no exact mapping exists the nearest model is used and the README SHALL say it is approximate.

#### Scenario: Two races
- **WHEN** two players of different races are in view
- **THEN** each is drawn with its own race's base model

#### Scenario: Gender
- **WHEN** two players of the same race and different genders are in view
- **THEN** each is drawn with the matching gender's model

#### Scenario: Unknown or missing
- **WHEN** a race or gender has no converted model
- **THEN** the current plain model is drawn and nothing is logged to the player

### Requirement: The mapping is established from the install
The race-to-model mapping SHALL come from the install's data, found by a recorded spike, and SHALL NOT be forced by a guessed numeric offset.

#### Scenario: Spike outcome
- **WHEN** the spike ends
- **THEN** design.md records where it looked and whether the mapping is exact, approximate or not found
