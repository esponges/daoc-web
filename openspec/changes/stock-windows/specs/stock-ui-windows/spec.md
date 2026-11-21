# Spec Delta

## ADDED Requirements

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
Each game window SHALL be built from its own stock template and filled through that template's controls, with no plain HTML panel drawn over a borrowed frame. The windows and templates SHALL be: Attributes `StatsAttributes`, Spells `stats_spells`, Skills and Styles `stats_spec_abil`, Merchant `Merchant`, Pet `MiniPet`, Vault `Vault`, Task `Quest`, Trainer `Train`, Healer and NPC talk `Interact`, Concentration `Concentration`, Cast bar `Timer` and Quiver `WindowQuiver`.

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
