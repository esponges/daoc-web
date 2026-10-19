# Spec Delta

## Purpose

Makes the compass agree with the game's own idea of north, checked against the zone data rather than assumed.

## ADDED Requirements

### Requirement: The compass agrees with the map
The compass SHALL show north in the direction the zone's own map calls north. This SHALL be checked against at least one zone exit whose direction from a known point follows from the zone data.

#### Scenario: Facing a known exit
- **WHEN** the camera faces a zone exit whose direction from the character is known from the zone data
- **THEN** the compass shows that direction in its middle

### Requirement: The heading is the same in every zone
The compass SHALL use one north for every outdoor zone and interior, so that a heading read in one zone means the same in the next.

#### Scenario: Across a border
- **WHEN** a character walks north across a zone border without turning
- **THEN** the compass shows north before and after the border
