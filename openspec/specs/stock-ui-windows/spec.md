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
