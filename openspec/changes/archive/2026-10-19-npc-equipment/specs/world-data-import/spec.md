## ADDED Requirements

### Requirement: NPC equipment is imported and linked to spawns
The importer SHALL store every `npcequipment` row (template id, slot, model, colour, effect, extension, emblem) and SHALL record each spawn's equipment template id, taken from the mob row and, when that is empty, from its NPC template.

#### Scenario: Counts are reported
- **WHEN** the importer finishes
- **THEN** it prints rows read and kept for the equipment table and exits non-zero if the table is missing or empty

#### Scenario: Mob without equipment
- **WHEN** a mob and its NPC template both have no equipment template id
- **THEN** the spawn is imported with no equipment and no error

#### Scenario: Unknown template id
- **WHEN** a spawn names an equipment template that has no rows
- **THEN** the spawn is imported with no equipment and the importer counts it in a warning total
