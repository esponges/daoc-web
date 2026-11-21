# Spec Delta

## ADDED Requirements

### Requirement: The pet's fight appears in the log
The owner's log SHALL show their pet's hits with damage, its misses, its spells and their damage, and the blows and spells that land on the pet. These lines SHALL use the game's wording ("Your returned commander hits the svartalf guard for 23 damage!", "The svartalf guard hits your returned commander for 12 damage!") and SHALL go to the main tab. The minions of a Bonedancer's commander SHALL be logged the same way.

#### Scenario: The pet fights
- **WHEN** the owner's pet hits a creature for 23 damage and the creature hits back for 12
- **THEN** the owner's log shows "Your <pet> hits the <creature> for 23 damage!" and "The <creature> hits your <pet> for 12 damage!"

#### Scenario: Someone else's pet
- **WHEN** another player's pet fights near the owner
- **THEN** none of that fight appears in the owner's log

#### Scenario: Out of sight
- **WHEN** the owner's pet fights a creature far beyond what the owner can see
- **THEN** the owner's log still shows the pet's blows and the blows it takes, and no other player is told of that fight
