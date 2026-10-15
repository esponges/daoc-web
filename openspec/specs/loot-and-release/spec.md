# loot-and-release Specification

## Purpose
TBD - created by archiving change stock-ui-rest. Update Purpose after archive.

## Requirements

### Requirement: Loot in a stock window
Taking items and money from a corpse SHALL use a window built on the stock window art, listing each item with its picture by kind, its name and its detail, with a Take all button, and closing when the character walks away.

#### Scenario: Loot a frog
- **WHEN** the character kills a creature and opens its corpse
- **THEN** the loot window lists the money and items, clicking one takes it, and Take all takes the rest

#### Scenario: Walking away
- **WHEN** the character moves beyond the loot distance
- **THEN** the window closes

### Requirement: Release after death
When the character dies the stock Release window SHALL appear with a timer and a release control. Pressing release SHALL send a release request, and the server SHALL stand a dead character up at their bind point at once and ignore the request from a living one; without it the server stands them up when its timer ends.

#### Scenario: Dying and releasing
- **WHEN** the character dies
- **THEN** the Release window shows, and pressing release returns the character to life where the server puts them

### Requirement: Fallback without converted data
Without the converted interface the built-in loot box and death banner SHALL be used.

#### Scenario: No converted UI
- **WHEN** the viewer starts without `web/data/ui/`
- **THEN** loot and death look and work as before
