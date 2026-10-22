# Spec Delta

## ADDED Requirements

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
