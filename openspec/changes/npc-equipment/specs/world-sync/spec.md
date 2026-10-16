## ADDED Requirements

### Requirement: NPC entities carry their worn equipment
A snapshot entity for an NPC SHALL include its worn equipment as a list of slot, model, colour and extension values, omitted when the NPC wears nothing. The field SHALL be additive, documented in `docs/protocol.md`, and ignored by clients that do not know it.

#### Scenario: Geared NPC
- **WHEN** a player is within range of an NPC whose spawn has equipment
- **THEN** its entity in the snapshot lists those pieces

#### Scenario: Ungeared NPC
- **WHEN** an NPC has no equipment
- **THEN** its entity has no equipment field

#### Scenario: Old client
- **WHEN** a client that predates the field receives the snapshot
- **THEN** it draws the NPC as before
