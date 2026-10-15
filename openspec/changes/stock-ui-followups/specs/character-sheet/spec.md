## ADDED Requirements

### Requirement: Character windows open from the menu bar
The attributes, styles, skills and spells buttons of the stock menu bar SHALL open stock windows built from their layouts, and SHALL close them on a second press.

#### Scenario: Open the attributes window
- **WHEN** the player presses ATTRIB. on the menu bar
- **THEN** the stock attributes window opens showing the character's level, class, race and attributes

#### Scenario: Close it again
- **WHEN** the player presses the same button again
- **THEN** the window closes

### Requirement: Windows show only what the server knows
A character window SHALL show a value only if the server has sent it, and SHALL leave a field blank rather than invent one.

#### Scenario: A value the server does not track
- **WHEN** a window has a field for something the server does not send
- **THEN** the field is empty

### Requirement: The server sends the data a window needs
Where a window needs a value the server holds and does not yet send, the server SHALL send it to its owner, and only to its owner, in an additive message.

#### Scenario: Another player's attributes
- **WHEN** a player is near another
- **THEN** the other player's attributes are not in any message sent to them

### Requirement: Windows without data say why
A menu button whose window cannot be shown (journal, map, warmap, and group while there are no groups) SHALL be greyed and its tooltip SHALL say why.

#### Scenario: Pointing at the journal button
- **WHEN** the pointer rests on the greyed JOURNAL button
- **THEN** a tooltip says the window is not available yet
