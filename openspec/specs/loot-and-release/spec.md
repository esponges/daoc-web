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

### Requirement: Loot by creature level
A kill SHALL pay coin and, with a chance that grows with the creature's level (22% at level 1, one point more per two levels, at most 40%), one item whose level lies within two levels of the creature's. Named creatures that have drops in the data SHALL keep those. The chance is marked provisional in the rule book until a source for 1.65's generators is adopted. Fitting the item to the creature's kind (hides and parts for animals) waits for item-kind data the import does not carry.

#### Scenario: Ordinary kills
- **WHEN** a character kills twenty level 5 creatures with no named drop
- **THEN** about a quarter of the kills (24%) leave an item, each within two levels of the creature

#### Scenario: A named creature
- **WHEN** a named creature with drops in the data is killed
- **THEN** its drops are rolled as before
