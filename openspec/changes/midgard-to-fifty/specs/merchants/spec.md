## ADDED Requirements

### Requirement: Gear is for sale up to level 50
For every five-level band from 31 to 50, some Midgard merchant SHALL sell each weapon kind, shield, and armour type that a Midgard class can use in that band. The stock SHALL come from the data's non-unique Midgard item templates of that level, or from the data's own merchant lists where they hold such gear. It SHALL be sold where the findings place it, or in Jordheim and the frontier towns when they place it nowhere. Status: set by the findings.

#### Scenario: Census bands
- **WHEN** `go run ./cmd/parity` runs after this change
- **THEN** the levelling gear table reports no gap in any band from 1-5 to 46-50

#### Scenario: Buying at 45
- **WHEN** a level 45 Warrior opens the merchant that sells chain armour for 41-45
- **THEN** a level 41 to 45 chain piece is offered and can be bought and worn

### Requirement: Added stock is recorded, not mixed into the data's lists
Stock that the data's merchant lists do not hold SHALL be added by this project's own table, kept under source control. It SHALL be marked as added by the project, with the reason, and the importer SHALL write it into the world store alongside the data's lists.

#### Scenario: A re-import
- **WHEN** the world store is re-imported
- **THEN** the added stock is present again, and the census can tell added stock from the data's
