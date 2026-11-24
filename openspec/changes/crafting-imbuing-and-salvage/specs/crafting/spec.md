## ADDED Requirements

### Requirement: Stations
A recipe that needs a forge, lathe or alchemy table SHALL be refused unless one is near.

#### Scenario: No forge
- **WHEN** a crafter starts a metal recipe away from any forge
- **THEN** it is refused with "You must be near a forge to make this."

### Requirement: Spellcraft
A Spellcrafter SHALL imbue crafted items with gems, up to the item's imbue points. Going past the points SHALL carry the rule book's chance of the item exploding.

#### Scenario: Within points
- **WHEN** a Spellcrafter imbues two gems within an item's points
- **THEN** both bonuses are on the item and nothing can explode

### Requirement: Salvage
Salvaging an item at the right station SHALL destroy it and return some of its materials.

#### Scenario: Salvage
- **WHEN** a crafter salvages a metal sword at a forge
- **THEN** the sword is gone and metal bars are in the pack
