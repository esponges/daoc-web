# zone-fidelity Specification

## Purpose

Defines what a converted outdoor zone must show in the viewer (terrain, ground detail, grass and placed scenery) and how the census reports and guards that per zone, so that "converted" means "looks right" and not just "files read".

## Requirements

### Requirement: Terrain draws in every converted zone
The viewer SHALL draw a converted zone's terrain whether or not the zone has ground detail layers, and the draw SHALL raise no WebGL error.

#### Scenario: Zone without ground detail
- **WHEN** the viewer opens a zone whose zone.json has no ground detail (for example zone 102, West Svealand)
- **THEN** the terrain is drawn from the colour atlas, lit and fogged like any other zone, and `gl.getError()` reports no error after a frame

#### Scenario: Zone with ground detail
- **WHEN** the viewer opens zone 100 (Vale of Mularn)
- **THEN** it renders as before this change

### Requirement: Ground detail converts from every reference tile format
zoneconv SHALL accept the zone's pre-blended reference tiles as either DDS or BMP when checking how the detail layers combine.

#### Scenario: Older zone with BMP tiles
- **WHEN** zoneconv converts zone 102, whose tex archive holds `tex*.bmp`
- **THEN** the zone gets ground detail and zoneconv prints the agreement it measured against those tiles

### Requirement: Ground detail finds frontier layer textures
zoneconv SHALL look up a frontier zone's layer textures in the frontier terrain-texture directory before the classic one.

#### Scenario: Frontier Odin's Gate
- **WHEN** zoneconv converts zone 167
- **THEN** layers such as `Odin_color` are found and the zone gets ground detail

### Requirement: Ground detail accepts masks of any square size
zoneconv SHALL resample square mask tiles of any size to the viewer's mask resolution, instead of skipping the zone.

#### Scenario: 64 px masks
- **WHEN** zoneconv converts zone 170, whose masks are 64 px
- **THEN** the zone gets ground detail

### Requirement: Weighted-average layer blends are drawn
When a zone's layers fit the game's own tiles best as a weighted average, zoneconv SHALL record that blend rule and the viewer SHALL draw the layers with it.

#### Scenario: Weighted zone
- **WHEN** zoneconv converts zone 108 (Raumarik)
- **THEN** zone.json carries ground detail marked as a weighted blend, and the viewer draws the layers normalised by their mask weights

#### Scenario: Painted zone unchanged
- **WHEN** zoneconv converts zone 100
- **THEN** its ground detail is marked as painted in order, and the output is byte-identical to that before this change

### Requirement: Grass converts from higher-resolution maps
zoneconv SHALL convert grass when the grass and density maps are an integer multiple of the heightmap grid, placing grass at the maps' own resolution.

#### Scenario: Shrouded Isles grass
- **WHEN** zoneconv converts zone 152, with 512 px grass maps and a 256-cell heightmap
- **THEN** the zone gets grass

#### Scenario: Zone without a grass map
- **WHEN** a zone's data archive has no grass map (for example zone 111, Uppland)
- **THEN** zoneconv skips grass and the census records "no grass map" as the reason

### Requirement: Tree clusters are placed as their trees
A placed model listed in the game's tree cluster table SHALL be converted as its base tree, drawn once at each of the cluster's offsets, rotated and scaled with the placement.

#### Scenario: Frontier forest
- **WHEN** propconv converts zone 170 (frontier Uppland)
- **THEN** placements of `Npintre1-sCL5.nif` and the other clusters appear as groups of trees, and the census counts them as converting

#### Scenario: Cluster with fewer offsets
- **WHEN** a cluster row has unused offset slots, written as zeros
- **THEN** only the listed offsets produce trees

### Requirement: Atlas layout is verified against the game's own tiles
zoneconv SHALL choose the colour atlas tile layout by agreement with the zone's pre-blended reference tiles when they exist, and SHALL report the agreement for the chosen layout and the best alternative.

#### Scenario: Flat zone
- **WHEN** zoneconv converts zone 107 (Vanern Swamp), where brightness barely tracks height
- **THEN** the layout is chosen by the reference tiles and printed with its agreement, and the terrain colours line up with the zone's shorelines in the viewer

#### Scenario: No reference tiles
- **WHEN** a zone has no reference tiles
- **THEN** zoneconv falls back to correlating brightness with height, and warns when that correlation is weak

### Requirement: Census reports per-zone fidelity
For every outdoor zone, the census SHALL report whether the terrain is drawable, the atlas layout and its agreement, whether ground detail converts (or why not), whether grass converts (or why not), and placements against those converting.

#### Scenario: Reason recorded
- **WHEN** a zone's ground detail cannot convert
- **THEN** its census row carries the reason in a `detail` field, and README.md lists it

#### Scenario: Deterministic output
- **WHEN** the census runs twice on the same install
- **THEN** census.json is byte-identical

### Requirement: Midgard zones are guarded against regression
A test SHALL fail if any Midgard outdoor zone present in the install loses terrain, ground detail, grass or converting placements, compared with the counts recorded in the census.

#### Scenario: Every Midgard zone complete
- **WHEN** the test runs against an install with zones 100–108, 111–113, 115, 116, 151–156, 158 and 167–170
- **THEN** each zone has drawable terrain and ground detail, each zone that has a grass map has grass, and every placement converts unless the census records a verified reason it cannot

#### Scenario: No install
- **WHEN** the test runs under `-short` or without a game install
- **THEN** it is skipped

### Requirement: Census reports interior zones
For every city and dungeon zone present, the census SHALL dry-run its layout as the converter reads it, and report the zone's kind, its placements against those converting, the models that do not convert and why, and whether its entrances and lights lie within its geometry.

#### Scenario: A dungeon row
- **WHEN** the census runs
- **THEN** zone 125's row gives kind dungeon, 196 placements (194 pieces and 2 props) with all converting, and its entrances and lights inside the geometry

#### Scenario: Deterministic output
- **WHEN** the census runs twice on the same install
- **THEN** census.json is byte-identical, interior rows included

### Requirement: Midgard interiors are guarded against regression
A test SHALL fail if Jordheim or any of the twelve Midgard dungeons present in the install loses a converting placement, or has an entrance or light outside its geometry, compared with the census baseline.

#### Scenario: Every Midgard interior complete
- **WHEN** the test runs against an install with zones 120, 125–129, 149, 150, 160, 161, 188, 189 and 226
- **THEN** every placement in each converts, unless the census records a verified reason it cannot, and every entrance and light lies within the geometry

#### Scenario: No install
- **WHEN** the test runs under `-short` or without a game install
- **THEN** it is skipped
