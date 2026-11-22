# game-messages Specification

## Purpose
TBD - created by archiving change stock-ui-rest. Update Purpose after archive.

## Requirements

### Requirement: Game messages appear in the chat window
Combat results, experience and level gains, loot and server refusals SHALL appear as lines in the stock chat window, coloured by kind, instead of in a separate log.

#### Scenario: A fight
- **WHEN** the character hits a creature and is hit back
- **THEN** both lines appear in the chat window, the character's hits and the creature's hits in different colours

#### Scenario: A refusal
- **WHEN** the server refuses an action with a reason
- **THEN** the reason appears in the chat window as an error line

### Requirement: Lines are sorted onto chat tabs
The chat window SHALL keep spoken lines and game messages apart on its tabs, so that a busy fight does not bury conversation.

#### Scenario: Reading the tabs
- **WHEN** the player selects the tab for spoken lines during a fight
- **THEN** combat lines are not shown there and still wait on the main tab

### Requirement: Fallback without converted data
Without the converted interface the built-in log SHALL be used as before.

#### Scenario: No converted UI
- **WHEN** the viewer starts without `web/data/ui/`
- **THEN** messages appear in the built-in log

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
