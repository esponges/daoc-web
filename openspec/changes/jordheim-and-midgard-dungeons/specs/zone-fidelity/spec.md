# Spec Delta

## ADDED Requirements

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
