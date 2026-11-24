# player-appearance Specification

## Purpose
What the server tells clients about a player's looks (race, gender and worn equipment) and how a client draws a player from that gear, live as it changes.

## Requirements

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

### Requirement: Appearance choices at creation
Character creation SHALL offer each choice the install provides for the chosen race and gender:
- face;
- hair style;
- hair colour;
- eye colour.

The server SHALL accept only values valid for that race and gender.

#### Scenario: Choosing hair
- **WHEN** a player creates a female Norse character with hair style 3 and colour 2
- **THEN** the character wears that hair in the creation preview and in the world

#### Scenario: Invalid choice
- **WHEN** a creation request names a face the race does not have
- **THEN** creation is refused with the reason

### Requirement: Everyone sees the same appearance
Other players SHALL see a character's chosen face, hair and colours.

#### Scenario: Seen by another
- **WHEN** Bjorn looks at Thora
- **THEN** he sees the hair and face she chose
