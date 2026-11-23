## ADDED Requirements

### Requirement: Quest talk
Talking to an NPC SHALL show, after its own text, the quests it offers to this player and any active quest step that ends with it. Delivering or finishing SHALL happen in the same talk window.

#### Scenario: Turning in
- **WHEN** a player whose current step is "bring the pelt to Svala" talks to Svala with the pelt
- **THEN** the talk window offers to hand it over, and doing so advances the quest
