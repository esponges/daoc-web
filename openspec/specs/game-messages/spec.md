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
