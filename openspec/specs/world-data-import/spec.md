# world-data-import Specification

## Purpose
TBD - created by archiving change multiplayer-server. Update Purpose after archive.

## Requirements

### Requirement: Community data is imported locally and never committed
The importer SHALL read a local checkout of the community database from a gitignored directory and write this project's own world store there, and the repository SHALL contain no rows, dumps or code from that database.

#### Scenario: Clean working tree
- **WHEN** the importer has run on a machine with the checkout
- **THEN** `git status` shows no community data, and the data directories are listed in `.gitignore`

#### Scenario: No checkout
- **WHEN** the server starts and the world store does not exist
- **THEN** it says how to fetch the dump and run the importer, and starts with no spawns instead of crashing

### Requirement: The importer reads SQL dumps without a database server
The importer SHALL parse the per-table SQL files directly and SHALL import the tables the game needs: spawns, NPC templates and equipment, loot, items, spells, styles, classes and specialisations, races, regions, jump points and start locations.

#### Scenario: Table counts
- **WHEN** the importer finishes
- **THEN** it prints rows read and rows kept per table and exits non-zero if a required table is missing or empty

#### Scenario: Unknown column
- **WHEN** a dump has a column the importer does not know
- **THEN** it ignores the column and reports it

### Requirement: Imported models resolve against the client install
Each imported NPC SHALL be matched to a figure the converters can draw, and those that cannot be matched SHALL be reported and left out of the world.

#### Scenario: Mularn creatures
- **WHEN** the importer finishes for zone 100
- **THEN** each kept spawn has a model that `charconv -model` has converted or can convert, and the unmatched count is printed with examples

### Requirement: Imported positions fit the client's zones
The importer SHALL convert region coordinates into this project's zone coordinates and drop spawns that fall outside the zone or far from the ground.

#### Scenario: Spawn on the ground
- **WHEN** a kept spawn is placed
- **THEN** its height is within a set distance of the derived heightmap at that point

#### Scenario: Offset not known
- **WHEN** the importer cannot establish the transform for a zone from the data
- **THEN** it reports the zone as not imported rather than guessing

### Requirement: Licence and scope are stated
The README SHALL state that the imported data is GPL-3.0, describes content owned by Electronic Arts, is not redistributed here, and that use is private.

#### Scenario: Reading the README
- **WHEN** someone reads the multiplayer section
- **THEN** it names the source repository, its licence and what must not be committed

### Requirement: Only creatures the client can draw are kept
The importer SHALL keep a spawn only if its model has a converted character, and SHALL report the models left out.

#### Scenario: Unconverted models
- **WHEN** the importer runs with some models unconverted
- **THEN** their spawns are dropped and the most common missing models are printed, and `-all-models` keeps them all for listing what to convert

### Requirement: Zone coordinates come from the zone table
The importer SHALL place a spawn in the zone whose region, offset and size contain it, with local units equal to region units minus the zone's offset times 8192.

#### Scenario: Mularn
- **WHEN** zone 100's spawns are imported
- **THEN** all but a few percent lie within 500 units of the derived ground

### Requirement: Class type and stock list are kept with each spawn
The importer SHALL keep each spawn's class type and the id of the stock list it names, and SHALL import the merchant stock table (list id, item id, page, slot). It SHALL print read and kept counts for that table and fail when the table is missing or empty.

#### Scenario: A merchant row
- **WHEN** a synthetic dump has a mob of class type `DOL.GS.GameMerchant` naming list `L1` and two stock rows for `L1`
- **THEN** the spawn row carries both values and the stock table holds two rows

#### Scenario: Missing stock table
- **WHEN** the merchant stock file is absent or has no rows
- **THEN** the import fails and says which table

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

### Requirement: A re-import can be staged while the server runs
When `world.db` is held by a running server, the importer SHALL write the new store to `world.db.next` and say so, instead of failing. On its next start the server SHALL check that the staged store opens and has the expected tables. If it does, the server SHALL move `world.db` to `world.db.prev` and the staged store into place. A staged store that fails the check SHALL be left in place, the server SHALL keep the old store, and the log SHALL say why.

#### Scenario: Import while running
- **WHEN** `dbimport` runs while the game server has `world.db` open
- **THEN** it writes `world.db.next`, prints that a restart will use it, and exits zero

#### Scenario: Restart picks it up
- **WHEN** the server is restarted with a valid `world.db.next` present
- **THEN** it starts on the new store, the old one is at `world.db.prev`, and `players.db` is untouched

#### Scenario: A broken staged store
- **WHEN** `world.db.next` is truncated
- **THEN** the server starts on the old `world.db`, leaves the staged file, and logs why it was refused

### Requirement: A template that replaces mob values wins
When a spawn's template is marked `ReplaceMobValues`, the importer SHALL take the spawn's name, aggro level and range, flags, speed (when set) and class type (when not the plain NPC) from the template, and keep the template's level list. The server SHALL pick one of those levels each time the creature spawns. Model, size and equipment stay the spawn's. The import SHALL print how many spawns were replaced. Status: reference (OpenDAoC-Core loads its data this way).

#### Scenario: A row the template corrects
- **WHEN** a spawn row says level 26 and aggressive, and its marked template says level `1` and aggro 0
- **THEN** the stored spawn is level 1 and passive, and is counted as having a level outside the template's

#### Scenario: A level list
- **WHEN** a marked template's level is `2;4;3`
- **THEN** the spawn keeps that list, and each time it spawns its level is one of 2, 3 and 4

#### Scenario: An unmarked template
- **WHEN** a spawn's template is not marked
- **THEN** the spawn's own name, level and aggression are kept

### Requirement: Spells and their links are imported
The importer SHALL store the spells that any kept NPC template, style or item names, with their type, target, range, cast time, recast delay, damage, damage type, value, duration, frequency, life-drain return, sub-spell and uninterruptible flag. It SHALL also store each template's spell list, the style-spell table, and each item's proc spell and proc chance. It SHALL print read and kept counts for each.

#### Scenario: A caster template
- **WHEN** a synthetic template lists spells `1001;1002`, and both are in the spell dump
- **THEN** the template's spell list holds both, and both spells are stored with their values

#### Scenario: A spell that is not in the dump
- **WHEN** a template names a spell id that the spell dump lacks
- **THEN** the link is dropped and counted in a warning total

### Requirement: Paths, start equipment and night spawns are imported
The importer SHALL store the path table and its points (path name, step, position, speed, wait). It SHALL also store the starting-equipment table split into one row per class, and whether each spawn is night-only by its class type (`NightSpawn`) or its brain. It SHALL print counts and SHALL fail when the starting-equipment or path table is missing or empty.

#### Scenario: A horse route
- **WHEN** a synthetic dump has a path `HS_A_B` with three points
- **THEN** the store holds the path with its three points in step order

#### Scenario: Starting equipment for several classes
- **WHEN** a starting-equipment row lists classes `35;31;22;`
- **THEN** the store holds that item once for each of classes 35, 31 and 22

#### Scenario: A night spawn
- **WHEN** a mob has class type `DOL.GS.NightSpawn`
- **THEN** its spawn row is marked night-only

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
