# Spec Delta

## Purpose

Measures how much of an installed copy of the game the pipeline can read and convert, per asset class, so "the whole world loads" is a number with every gap named rather than a claim.

## ADDED Requirements

### Requirement: Census covers every asset class in the install
The census SHALL enumerate, from a given install directory, every item of each asset class: zones listed in `zones.dat`, outdoor scenery models, dungeon pieces, city block models, character figures, animations and the textures those reference. It SHALL include the `zones/`, `frontiers/` and `Newtowns/` trees, not only `zones/`.

#### Scenario: All locations are counted
- **WHEN** the census runs against an install that has scenery models in `zones/Nifs`, `frontiers/NIFS` and `Newtowns/zones/Nifs`
- **THEN** the scenery total includes the models from all three locations, and the report lists each location's count

#### Scenario: Zones are classified by type
- **WHEN** the census reads `zones.dat`
- **THEN** each zone is counted under its declared type (outdoor, city, dungeon or other) with its number and name, and zones whose data folder is missing are reported as missing rather than omitted

### Requirement: Each item is tested by actually reading it
For each item the census SHALL attempt the same read the pipeline performs (parsing a model, an animation, a texture; converting an outdoor zone's terrain and placements) and record pass or fail. An item SHALL NOT be counted as passing on the basis of its file merely existing.

#### Scenario: A model that fails to parse is a failure
- **WHEN** a scenery model contains a block type the reader does not support
- **THEN** that model is recorded as failed, with the cause, and does not count towards the pass total

#### Scenario: An unreadable item does not stop the census
- **WHEN** one item causes a read error or contains corrupt data
- **THEN** the census records that item's failure and continues with the rest

### Requirement: Failures are grouped by cause
The report SHALL group failures within each asset class by cause, giving each cause a count and example item names, ordered by count, so the most widespread gap is first.

#### Scenario: Unsupported block types are aggregated
- **WHEN** 64 models fail because of the same unsupported block type
- **THEN** the report shows that block type once with a count of 64 and some example file names, rather than 64 separate lines

### Requirement: Report is reproducible and contains no game content
The report SHALL be generated entirely from the install and the pipeline, never edited by hand, and SHALL contain only names, counts, rates and failure causes. It MUST NOT contain game assets or data extracted from them beyond file names.

#### Scenario: Regenerating gives the same report
- **WHEN** the census runs twice against the same install and the same code
- **THEN** the two reports are identical

#### Scenario: Report is safe to commit
- **WHEN** the report is written into the repository
- **THEN** it holds file names, counts and causes only, with no geometry, textures, tables or text taken from game files

### Requirement: Outdoor zones report conversion status
For every outdoor zone the census SHALL record whether its terrain and its scenery placements convert, and for scenery the fraction of placements whose model converts.

#### Scenario: A zone with an unconvertible model
- **WHEN** an outdoor zone places 1,000 fixtures and 10 of them use a model that fails to parse
- **THEN** the zone is reported with 990 of 1,000 placements converting and the failing models named

### Requirement: Parse rates cannot silently regress
A check SHALL compare each asset class's pass count with the last recorded census and fail when any class's pass count falls. The check SHALL be skipped, with a message saying why, when no install is available.

#### Scenario: A change breaks parsing
- **WHEN** a code change causes previously passing figures to fail
- **THEN** the check fails, naming the asset class and the old and new pass counts

#### Scenario: No install present
- **WHEN** the check runs on a machine without the game installed
- **THEN** it is skipped with a message naming the install path it looked for
