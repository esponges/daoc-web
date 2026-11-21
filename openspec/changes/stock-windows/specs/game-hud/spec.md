# Spec Delta

## MODIFIED Requirements

### Requirement: Chat window
The HUD SHALL show chat in the stock chat window, send typed lines over the existing chat message, and colour lines by channel. The window's tabs SHALL filter what it shows: Main SHALL show every line, including combat, spells and system lines, and the other tabs SHALL show their own channels. The window SHALL keep at least 200 earlier lines that the player can scroll back through with its scrollbar or the mouse wheel. It SHALL be resizable and movable, and start large enough to show at least twelve lines.

#### Scenario: Say
- **WHEN** a player types a line and presses Enter
- **THEN** the server broadcasts it and every nearby chat window shows it

#### Scenario: Scrolling back
- **WHEN** a fight has pushed earlier lines out of view and the player scrolls up
- **THEN** the earlier lines show again, and a new line does not jump the view back down until the player scrolls to the bottom

#### Scenario: Tabs
- **WHEN** the player selects the Chat tab during a fight
- **THEN** only spoken lines show, and selecting Main again shows the combat lines too

### Requirement: Style and quick bar
The HUD SHALL show the character's styles and spells in the stock quickbar of ten slots, drawn from the skin's art, with each icon drawn by its weapon line or spell kind. A slot SHALL be used on click or its key, and its tooltip SHALL show the opener requirement or the spell's cost and cast time. The bar SHALL shade a slot that cannot be used now, show a style the character cannot pay for in red, show a recharge shade while waiting, and rebuild when the weapon changes.

#### Scenario: Use a style
- **WHEN** the player clicks a style icon
- **THEN** the style message is sent and the icon shows a recast shade until the server allows it again

#### Scenario: Cast from a slot
- **WHEN** the player presses the key of a slot holding a spell
- **THEN** the spell is cast at the current target and the slot shades until the spell can be cast again
