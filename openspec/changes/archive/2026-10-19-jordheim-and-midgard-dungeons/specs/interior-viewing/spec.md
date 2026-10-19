# Spec Delta

## Purpose

Presents a converted city or dungeon zone in the viewer: how it is opened, where the character starts, what it stands on, and how it is lit without the outdoor sun.

## ADDED Requirements

### Requirement: Interiors open like any zone
The viewer SHALL open a converted city or dungeon zone from `?zone=N` and from the zone picker, drawing its geometry without terrain, water, grass or sky.

#### Scenario: Open Spindelhalla
- **WHEN** the viewer is opened with `?zone=125`
- **THEN** the dungeon's pieces and props are drawn, no terrain or sky is drawn, and the first frame raises no WebGL error

#### Scenario: Picker
- **WHEN** the zone picker lists zones
- **THEN** Jordheim and the Midgard dungeons appear among them, marked as city or dungeon

### Requirement: The character starts at an entrance
The viewer SHALL start the character at an entrance of the interior, from the trigger boxes in its manifest, standing on the floor and facing into the interior.

#### Scenario: Entering Spindelhalla
- **WHEN** zone 125 opens
- **THEN** the character stands on the cave floor near the `ExittoMyrk` entrance, inside the geometry rather than above, below or beside it

### Requirement: The character walks on the geometry
In an interior, the viewer SHALL find the ground under the character by casting down onto the interior's triangles, taking the nearest floor below a step's height above the character's feet.

#### Scenario: Walking down a ramp
- **WHEN** the character walks along a sloping passage
- **THEN** its feet follow the floor, and a ceiling or the floor of an upper level is never taken as the ground

#### Scenario: Off the edge
- **WHEN** no triangle lies below the character
- **THEN** it keeps its last ground height rather than falling without end

### Requirement: Interiors are lit from within
The viewer SHALL light an interior from the models' baked shading, an ambient level and the zone's point lights, and SHALL NOT apply the outdoor sun or its shadows inside.

#### Scenario: A lit dungeon
- **WHEN** zone 125 is viewed near one of its lights
- **THEN** surfaces near the light are brighter than those far from any light, and nothing is lit from the sky's direction

#### Scenario: Outdoor zones unchanged
- **WHEN** zone 100 is viewed
- **THEN** it is lit and shadowed as before this change
