# Spec Delta

## MODIFIED Requirements

### Requirement: Stock menu bar, compass and clock
The HUD SHALL show the stock menu bar, compass and clock windows, with the compass turning with the camera and the clock showing the server's game time of day (`world-time`). A menu button whose window exists (INV., ATTRIB., SKILLS, STYLES, GROUP) SHALL open it. One whose window does not (SPELLS, JOURNAL, MAP, WARMAP) SHALL be greyed, with a tooltip saying why.

#### Scenario: Turning
- **WHEN** the character turns to face another direction
- **THEN** the compass shows the new heading

#### Scenario: Menu buttons
- **WHEN** the player presses the inventory button on the menu bar
- **THEN** the inventory opens as it does on the key

#### Scenario: A button with no window
- **WHEN** the player points at the greyed SPELLS button
- **THEN** a tooltip says a Warrior has no spells and spellcasting is not built yet

#### Scenario: The clock
- **WHEN** the server's game time is 18:30
- **THEN** the stock clock shows 18:30, not the player's local time
