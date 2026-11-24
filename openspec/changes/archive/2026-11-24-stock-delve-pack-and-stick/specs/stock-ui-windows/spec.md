## ADDED Requirements

### Requirement: Right-click delves
Right-clicking an item, spell or style wherever it is shown SHALL ask the server for its details and write them to the chat log. A piece of gear's name SHALL be drawn in its con colour to the player. Shift+Right-click on a quickbar or spell bar slot SHALL clear it; a plain right-click SHALL NOT.

#### Scenario: Delving a sword in the pack
- **WHEN** a level 10 player right-clicks a level 12 sword in the pack
- **THEN** the log shows the sword's name in its con colour, its level, DPS, speed and damage type

#### Scenario: Clearing a quickbar slot
- **WHEN** the player Shift+Right-clicks a filled quickbar slot
- **THEN** the slot is cleared, and a plain right-click on it delves instead

### Requirement: Items are drawn by kind
Each item SHALL be drawn with an icon chosen by hand from the install's sheets for its kind (weapon line or object type, armour slot, shield size, ammunition, consumable, material), never by an offset applied to the imported icon number.

#### Scenario: A bow and arrows
- **WHEN** the pack holds a bow and a stack of arrows
- **THEN** they draw different icons, neither of them the hammer or the generic item

### Requirement: The pack is the stock list
With the converted interface, a pack page SHALL show its eight slots as the paper doll art draws them: one column of rows, each with a small icon and the item's name and count.

#### Scenario: A page with three items
- **WHEN** the first page holds three items
- **THEN** they fill the first three rows from the top, each with its name beside its icon
