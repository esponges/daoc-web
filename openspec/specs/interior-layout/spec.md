# interior-layout Specification

## Purpose
Reads a city or dungeon zone, which has no heightmap and is assembled from model pieces, into baked scenery in zone coordinates, along with the entrances and lights the viewer needs to present it.

## Requirements

### Requirement: Dungeon pieces are placed from the layout
For a dungeon zone, the converter SHALL place each `dungeon.place` row's piece (indexed into `dungeon.chunk`) at its position and axis-angle rotation, with pieces found in the dungeon piece directories.

#### Scenario: Spindelhalla
- **WHEN** the converter runs on zone 125
- **THEN** it writes 194 placed pieces from 26 piece models, and every piece converts

#### Scenario: Unknown piece index
- **WHEN** a placement names an index past the end of `dungeon.chunk`
- **THEN** that placement is skipped and counted, and the rest convert

### Requirement: Dungeon props are placed in their frame
The converter SHALL place each `dungeon.prop` row's model (indexed into `dungeon.chunk`) with its position, rotation and scale, in the frame the layout data establishes, whether dungeon space or the piece the row names.

#### Scenario: Abandoned Mines
- **WHEN** the converter runs on zone 226
- **THEN** its 707 props are written, and each one lies within the bounds of the placed pieces

### Requirement: City blocks are drawn in place
For a city zone, the converter SHALL convert every block `city.csv` lists, keeping each block's own coordinates, since the blocks are authored in a shared city space.

#### Scenario: Jordheim
- **WHEN** the converter runs on zone 120
- **THEN** all 31 blocks convert, none of them skipped for non-finite data

### Requirement: Interiors are in zone coordinates
The converter SHALL express an interior's geometry in the same zone coordinates its `ZONEJUMP.csv` entrances and `lights.csv` lights use, by a transform established from the data and recorded in the zone's manifest.

#### Scenario: Exits lie in the geometry
- **WHEN** any converted Midgard interior is checked
- **THEN** the centre of every `ZONEJUMP.csv` trigger box lies within the converted geometry's bounds, and so does every light from `lights.csv`

### Requirement: Entrances and lights are recorded
The converter SHALL record each interior's entrances (name and trigger box from `ZONEJUMP.csv`) and its point lights (position and the type column from `lights.csv`) in the zone's manifest.

#### Scenario: Spindelhalla's exit
- **WHEN** zone 125 is converted
- **THEN** its manifest lists the `ExittoMyrk` entrance and 82 lights

### Requirement: Interior conversion is checked like outdoor conversion
`worldconv` SHALL convert the Midgard interiors along with the outdoor zones, and SHALL fail an interior whose written placements differ from the count the census reports converting. `zones.json` SHALL give each zone's kind.

#### Scenario: worldconv run
- **WHEN** `go run ./cmd/worldconv` runs
- **THEN** Jordheim and the twelve Midgard dungeons are written, each with as many placements as the census counts, and `zones.json` marks them city and dungeon
