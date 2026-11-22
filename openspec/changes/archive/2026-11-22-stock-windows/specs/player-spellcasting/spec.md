# Spec Delta

## MODIFIED Requirements

### Requirement: The SPELLS window and quickbar
The SPELLS window SHALL be the stock `stats_spells` window, listing the spellbook in its tree, with one heading per spell line and each spell's icon by kind, name and level. A spell SHALL be castable from the window and placeable on the quickbar by dragging, and cast from there. The cast bar SHALL be the stock `Timer` window, showing the spell's name and its progress while the spell is cast. Without the converted interface the built-in panels SHALL list and cast spells.

#### Scenario: Cast from the bar
- **WHEN** the player drags a spell onto the bar and presses its key
- **THEN** the spell is cast at the current target

#### Scenario: The spell tree
- **WHEN** a Bonedancer opens SPELLS
- **THEN** the stock window shows headings for Bone Army, Darkness and Suppression, and expanding one lists its spells with their icons

#### Scenario: The cast bar
- **WHEN** the player begins a three-second cast
- **THEN** the stock Timer window shows the spell's name and fills over three seconds, and hides when the spell lands or is interrupted

#### Scenario: No converted UI
- **WHEN** `web/data/ui/` is absent
- **THEN** the built-in panel still lists the spells and casts them
