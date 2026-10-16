## ADDED Requirements

### Requirement: The server sends a player's race, gender and equipment
Player entities in snapshots SHALL carry the player's race, gender and worn equipment as additive fields, in the same form NPC equipment uses.

#### Scenario: Remote player in range
- **WHEN** a player comes within range of another
- **THEN** the snapshot entity carries that player's race, gender and equipment

#### Scenario: Older client
- **WHEN** a client does not know the new fields
- **THEN** it ignores them and still plays

### Requirement: Players are drawn from their gear
The client SHALL draw the local player and remote players with the pieces it can resolve under `npc-gear-mapping`, and SHALL update the look when equipment changes. A piece it cannot resolve SHALL be skipped without error, and with no lookup files players SHALL draw with the baked outfit as before.

#### Scenario: Equipping a weapon
- **WHEN** the local player equips a weapon that resolves
- **THEN** it appears in their hand and a nearby player sees it

#### Scenario: Unresolved piece
- **WHEN** a worn piece cannot be resolved
- **THEN** that slot keeps the base model and the skipped count rises

#### Scenario: Lookup absent
- **WHEN** the generated lookup files are missing
- **THEN** players draw with the baked outfit
