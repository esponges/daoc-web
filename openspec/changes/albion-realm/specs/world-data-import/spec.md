## ADDED Requirements

### Requirement: Albion regions imported
`dbimport` SHALL import the spawns, merchants, trainers, guards and items of Albion's regions (classic Albion, Avalon, Camelot and Albion's dungeons), with the same not-classic filtering as Midgard.

#### Scenario: Population
- **WHEN** `dbimport` runs and the server starts
- **THEN** the creatures of Albion's starting zones are in the world
