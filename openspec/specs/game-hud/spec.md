# game-hud Specification

## Purpose
TBD - created by archiving change stock-ui. Update Purpose after archive.

## Requirements

### Requirement: Chat window
The HUD SHALL show chat in the stock chat window, send typed lines over the existing chat message, and colour lines by channel.

#### Scenario: Say
- **WHEN** a player types a line and presses Enter
- **THEN** the server broadcasts it and every nearby chat window shows it

### Requirement: Target and health bars
The HUD SHALL show the player's own health, endurance and mana and the target's name, level colour and health, updated from snapshots.

#### Scenario: Taking damage
- **WHEN** the player is hit
- **THEN** the health bar drops to the server's value

### Requirement: Style and quick bar
The HUD SHALL show the character's styles as stock icons in a bar, drawn by the style's weapon line, use them on click or key, show the opener requirement in the tooltip, shade a style that cannot be used now and show a style the character cannot pay for in red, and rebuild when the weapon changes.

#### Scenario: Use a style
- **WHEN** the player clicks a style icon
- **THEN** the style message is sent and the icon shows a recast shade until the server allows it again

### Requirement: Inventory and gear
The HUD SHALL show the pack and the equipped items over the stock paper doll, the pack as five pages of eight, with a picture for each item drawn by what it is, move items by dragging, and show the server's refusal reason when a change is refused.

#### Scenario: Equip by drag
- **WHEN** the player drags a hammer onto the weapon slot
- **THEN** the equip message is sent, and on success the slot shows the hammer

#### Scenario: Refused item
- **WHEN** the server refuses an equip
- **THEN** the item stays where it was and the reason is shown

### Requirement: No protocol change
The HUD SHALL work over the existing protocol messages.

#### Scenario: Old and new client
- **WHEN** either UI connects
- **THEN** the server behaves identically

### Requirement: Experience bar
The HUD SHALL show the character's level and progress to the next level in the stock experience window: ten coarse marks for each tenth of the level and a fine bar for progress within the tenth being filled.

#### Scenario: Gaining experience
- **WHEN** the server reports more experience for the character
- **THEN** the fine bar fills and, after each tenth, a coarse mark is added
