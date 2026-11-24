# stock-ui-windows Specification

## Purpose
TBD - created by archiving change stock-ui. Update Purpose after archive.

## Requirements

### Requirement: Windows are built from converted layouts
The client SHALL build a window from its converted layout, drawing frames, buttons, image areas and text at the sizes and positions in the data, without per-window drawing code for what the layout describes.

#### Scenario: Open a window
- **WHEN** the client opens a window by id
- **THEN** its elements appear where the layout puts them, using the stock textures

### Requirement: Text uses the stock fonts
Text SHALL be drawn with the bitmap or TrueType fonts the layout names, in the colours it names.

#### Scenario: Chat line
- **WHEN** a chat line is drawn in the chat window
- **THEN** it uses the window's font and the colour of its channel

### Requirement: Buttons show their states
A button SHALL show its normal, highlit and pressed looks and report clicks to the HUD.

#### Scenario: Pressing a button
- **WHEN** the pointer presses and releases a button
- **THEN** it shows the pressed look while held and the HUD receives one click

### Requirement: Fallback without converted data
When `web/data/ui/` is absent the client SHALL use the current built-in UI and say so on the console.

#### Scenario: No converted UI
- **WHEN** the viewer starts without `web/data/ui/`
- **THEN** the built-in UI works as it does today

### Requirement: Every stock control type is drawn
The window builder SHALL draw every control type the converted layouts use, each in its template's own art, font and colours: list boxes, tree controls, checkboxes, tabs, icons, status icons, horizontal and vertical resize images, image buttons, text areas, sliders, combo boxes and scrollbars. A control type the builder cannot draw SHALL be reported on the console once per type, not silently skipped.

#### Scenario: A list box
- **WHEN** a window whose layout has a list box is given rows
- **THEN** the rows are drawn in the list template's font and row art, a scrollbar appears when they overflow, and clicking a row selects it

#### Scenario: A tree control
- **WHEN** a tree control is given lines grouped under headings
- **THEN** each heading expands and collapses with its check art and each line shows its small icon and text

#### Scenario: An unknown control
- **WHEN** a layout holds a control type the builder does not know
- **THEN** the console names that type once and the rest of the window still draws

### Requirement: Game windows are built from their own templates
Each game window SHALL be built from its own stock template and filled through that template's controls, with no plain HTML panel drawn over a borrowed frame. The windows and templates SHALL be: Attributes `StatsAttributes`, Spells `stats_spells`, Skills `stats_spec_abil`, Styles `stats_combat`, Merchant `Merchant`, Pet `MiniPet`, Vault `Vault`, Task `Quest`, Trainer `Train`, Healer and NPC talk `Interact`, Concentration `Concentration`, Cast bar `Timer` and Quiver `WindowQuiver`.

#### Scenario: Opening the trainer
- **WHEN** the player right-clicks their trainer with the converted interface present
- **THEN** the stock Train window opens with the character's specs in its list, the points to spend in its label, and its train button working

#### Scenario: Talking to a healer
- **WHEN** the player right-clicks a healer
- **THEN** the stock Interact window shows the healer's text

### Requirement: Client-drawn windows use the skin's art
A window the game's client draws without an XML layout, such as the quickbar, SHALL be built from the skin's own textures, not from colours and shapes of our own.

#### Scenario: The quickbar
- **WHEN** the game starts with the converted interface present
- **THEN** the quickbar's frame, slots and bar number are drawn from the skin's textures

### Requirement: Windows move, close and remember where they were
A stock window whose template allows moving SHALL move by dragging its title area, and one with a close button SHALL close when it is pressed. The client SHALL remember each window's position in the browser and restore it in the next session, keeping the window inside the screen.

#### Scenario: Moving the spells window
- **WHEN** the player drags the Spells window to a new place, closes it and reloads the page
- **THEN** the window opens at the place it was dragged to

#### Scenario: A smaller screen
- **WHEN** a remembered position lies outside a smaller browser window
- **THEN** the window opens moved back inside the screen

### Requirement: The Command window
The client SHALL build the stock Command window from its template. SIT, SPRINT, ATTACK, PRAY and RELEASE SHALL send their actions, and FOLLOW, STICK, FACE and BONUSES SHALL be greyed with a tooltip saying why. Without the converted interface, plain buttons SHALL offer sit and sprint.

#### Scenario: Sit from the window
- **WHEN** the player presses SIT in the Command window
- **THEN** they sit, and pressing it again stands them up

#### Scenario: A button not built
- **WHEN** the player hovers FOLLOW
- **THEN** it is greyed and its tooltip says following is not built yet

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
