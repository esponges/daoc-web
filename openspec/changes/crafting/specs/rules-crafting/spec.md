## ADDED Requirements

### Requirement: Success chance
The chance to succeed SHALL depend on the difference between the crafter's skill and the recipe's level. Status: documented where a 1.65 source gives the curve, provisional otherwise.

#### Scenario: Orange recipe
- **WHEN** a crafter's skill is 10 below a recipe's level
- **THEN** the success chance is the rule book's figure for that gap, and below the chance for an equal-level recipe

### Requirement: Quality
A successful craft SHALL yield a quality from 94 to 100 percent with the rule book's distribution. 100 percent is a masterpiece.

#### Scenario: Distribution
- **WHEN** 10,000 crafts of an equal-level recipe are simulated
- **THEN** the qualities match the rule book's table within tolerance

### Requirement: Craft time
Craft time SHALL follow the rule book's formula from the recipe level and material count.

#### Scenario: Time
- **WHEN** a level 500 recipe with 20 materials is crafted
- **THEN** it takes the rule book's figure in seconds

### Requirement: Imbue points and overcharge
An item's imbue points SHALL come from its level and quality. Each gem SHALL cost points by its bonus. Overcharge SHALL carry an explosion chance by how far past the points it goes and the crafter's spellcraft skill.

#### Scenario: Overcharge by 2
- **WHEN** a Spellcrafter overcharges an item by 2 points
- **THEN** the explosion chance is the rule book's figure for 2 points at that skill

### Requirement: Registered and tested
Each crafting rule SHALL have table tests per scenario and a registry entry with sources and status.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs
- **THEN** the crafting rules are listed
