# Spec Delta

## Purpose

Defines what a new character is given from the imported starting-equipment table, and the check that Midgard merchants sell what a Warrior needs at every level band.

## ADDED Requirements

### Requirement: A new character gets their class's starting equipment
When a character is created, the server SHALL give them every item the imported starting-equipment table lists for their class, equipped in the item's slot when that slot is free and the class may use it, otherwise placed in the pack.

#### Scenario: A new Warrior
- **WHEN** a Midgard Warrior is created
- **THEN** they wear the bronze stelskodd armour pieces, hold the training axe and the small training shield the table lists for class 22, and nothing else is added

#### Scenario: Two items for one slot
- **WHEN** the table lists two weapons for a class
- **THEN** the first by table order is equipped and the other is in the pack

### Requirement: Missing starting items are reported, not invented
If a listed item is not in the item table, the server SHALL skip it and log the class and item name. If a class has no rows, its characters SHALL start with nothing, and the log SHALL say so. No item SHALL be chosen by any other rule.

#### Scenario: An unknown template
- **WHEN** a row names an item id the item table lacks
- **THEN** the character is created without it and the log names the missing id

### Requirement: Existing characters keep what they have
The change of starting equipment SHALL apply only to characters created after it. No item SHALL be added to or taken from an existing character.

#### Scenario: A character from before the change
- **WHEN** a character created with the old kit logs in
- **THEN** their inventory is unchanged

### Requirement: Levelling gear is checked against the merchants
A check SHALL list, for each level band 1–5, 6–10 and so on to 46–50, the Warrior weapons (axe, hammer, sword, one- and two-handed) and armour (studded below level 10, chain from it) that Midgard merchants sell. It SHALL report each band and kind that no merchant sells. Gaps SHALL be reported, not filled.

#### Scenario: The report
- **WHEN** the parity census runs
- **THEN** its report has a levelling-gear table with each band and kind and the number of merchants selling it, and missing cells are marked

#### Scenario: Nothing is made up
- **WHEN** a band has no chain armour for sale
- **THEN** the report says so and no item or merchant row is added
