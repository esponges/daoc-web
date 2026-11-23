## ADDED Requirements

### Requirement: No silent skips
The server SHALL spawn every placement whose kind is implemented or "behaviour unsourced". A placement it skips SHALL be counted with its kind's reason in the parity census.

#### Scenario: Myling camps exist
- **WHEN** a player visits a Myling camp in its zone
- **THEN** the Mylings are there

### Requirement: Recharger
A Recharger SHALL restore the charges of an item with charges, for a price set in the rule book.

#### Scenario: Recharge
- **WHEN** a player hands a Recharger an item with 0 of 10 charges and pays
- **THEN** the item has 10 charges and the price is taken
