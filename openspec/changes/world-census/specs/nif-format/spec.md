# Spec Delta

## Purpose

Defines which NIF content the reader supports and how it reports what it does not, so model coverage is explicit and gaps are diagnosable rather than silent.

## ADDED Requirements

### Requirement: Unsupported content is reported precisely
When a model contains a block type the reader does not support, reading SHALL fail with an error naming the block type, its index and its byte offset, and SHALL NOT return partially parsed content as if it were complete.

#### Scenario: Unknown block type
- **WHEN** a model contains a block type with no reader
- **THEN** the read fails with an error that names that block type and where it occurs

### Requirement: Corrupt data is distinguished from unsupported data
The reader SHALL report malformed data (non-finite numbers where geometry is expected, implausible counts or string lengths) as corruption, with a different cause from an unsupported block type.

#### Scenario: Non-finite geometry
- **WHEN** a model's vertex data contains a non-finite float
- **THEN** the read fails with a cause identifying corrupt data, not an unsupported type

### Requirement: Effect and controller blocks found in scenery are read
The reader SHALL read the block types found in the install's scenery and dungeon models so the files containing them parse completely: at least `NiTextureEffect`, `NiGravity`, `NiTextureTransformController`, `NiGeomMorpherController`, `NiPixelData`, `NiAlphaController` and `NiMaterialColorController`.

#### Scenario: Model with a texture effect parses
- **WHEN** a scenery model containing an `NiTextureEffect` is read
- **THEN** it parses to the end of the file with every block accounted for, and its geometry converts

#### Scenario: Coverage rises accordingly
- **WHEN** the census runs after these block types are supported
- **THEN** no outdoor scenery model fails because of any of them

### Requirement: Particle gravity affects particles
Particle systems whose modifier chain includes `NiGravity` SHALL have their particles accelerated by that gravity's force and direction when drawn.

#### Scenario: Falling sparks
- **WHEN** an emitter has a gravity modifier pulling down
- **THEN** its particles' paths curve downward over their lifetime instead of travelling in straight lines

### Requirement: Embedded textures are usable
A texture stored inside a model (`NiPixelData`) rather than as an external file SHALL be extracted and used as that model's texture.

#### Scenario: Model with an internal texture
- **WHEN** a scenery model's texture is embedded in the file
- **THEN** the converted model draws with that texture rather than untextured
