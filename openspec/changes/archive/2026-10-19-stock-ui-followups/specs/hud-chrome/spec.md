## MODIFIED Requirements

### Requirement: Stock menu bar, compass and clock
The HUD SHALL show the stock menu bar, compass and clock windows, with the compass turning with the camera and the clock showing a time. A menu button whose window exists SHALL open it; one whose window does not SHALL be greyed with a tooltip saying why.

#### Scenario: Turning
- **WHEN** the character turns to face another direction
- **THEN** the compass shows the new heading

#### Scenario: Menu buttons
- **WHEN** the player presses the inventory button on the menu bar
- **THEN** the inventory opens as it does on the key

#### Scenario: A button with no window
- **WHEN** the player points at a greyed button
- **THEN** a tooltip says why it is not available
