## ADDED Requirements

### Requirement: Spell columns for concentration and resurrection
The importer SHALL add these spell columns: concentration cost, resurrection health and power percent, and focus flag. A world store without them SHALL load with no concentration spells and no resurrection, refusing those spells as unsupported.

#### Scenario: Re-imported store
- **WHEN** the store is re-imported
- **THEN** a Shaman's strength/constitution buff loads with its concentration cost, and a Healer's resurrection with its health and power percent

#### Scenario: Old store
- **WHEN** the server loads a store imported before this change
- **THEN** it starts, and casting a resurrection is refused as unsupported

### Requirement: Npc templates carry what a pet needs
The importer SHALL keep, for each npc template:
- its stats;
- its body type, so charm can tell an animal;
- its spells and styles;
- its left-hand chance.

A summon spell's template SHALL resolve to one of these rows.

#### Scenario: A summoned pet's template
- **WHEN** the server loads a Spiritmaster's summon spell
- **THEN** the template it names is found, with its model, level range, stats and body type

### Requirement: Data quests are counted and marked not classic
The importer and the census SHALL count Midgard's `dataquest` rows. Those whose name or kind marks them as "XP Item" turn-ins SHALL be marked not classic, with the reason, and SHALL never be offered in play.

#### Scenario: Census
- **WHEN** the census runs
- **THEN** the quests line reports all 81 Midgard quests as not classic, with the reason "XP item turn-in, a later freeshard addition"

### Requirement: Poison items keep their charges
The importer SHALL keep each item's charges and the spell it carries, so poisons and other charged items load whole.

#### Scenario: A poison item
- **WHEN** a Midgard poison item is loaded
- **THEN** its spell and charges are known to the server
