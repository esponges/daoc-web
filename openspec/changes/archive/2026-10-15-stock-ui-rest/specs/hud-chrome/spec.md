## ADDED Requirements

### Requirement: Developer information is hidden by default
The panel of zone, grid, triangle and frame figures and the key help SHALL be hidden until asked for with a key, and SHALL NOT take space on a normal screen.

#### Scenario: A new session
- **WHEN** a player enters the game
- **THEN** no developer panel is shown, and pressing the key shows and hides it

### Requirement: Stock menu bar, compass and clock
The HUD SHALL show the stock menu bar, compass and clock windows, with the compass turning with the camera and the clock showing the game's time of day.

#### Scenario: Turning
- **WHEN** the character turns to face another direction
- **THEN** the compass shows the new heading

#### Scenario: Menu buttons
- **WHEN** the player presses the inventory button on the menu bar
- **THEN** the inventory opens as it does on the key

### Requirement: Zone picker in the stock look
Choosing another zone SHALL use stock window art and fonts.

#### Scenario: Pick a zone
- **WHEN** the player picks a converted zone
- **THEN** the zone loads as before

### Requirement: Stock fonts and colours
Name labels over heads, the login page and any remaining page text SHALL use the stock fonts and the con colours of the game.

#### Scenario: A creature name
- **WHEN** a creature is in view
- **THEN** its name is drawn in a stock font in its con colour

### Requirement: Fallback without converted data
Without the converted interface the built-in panels, labels and zone picker SHALL be used.

#### Scenario: No converted UI
- **WHEN** the viewer starts without `web/data/ui/`
- **THEN** it looks and works as before
