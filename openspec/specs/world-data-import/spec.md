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
