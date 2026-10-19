# npc-gear-mapping Specification

## Purpose
How an imported item model number becomes a converted model from this install: weapons through the install's item table, armour only where a real mapping exists, and otherwise by the armour type the database gives each piece.

## Requirements

### Requirement: Held items resolve through the install's item table
The client SHALL draw an NPC's weapon or shield only when its model number resolves to an item model in the install's `items.csv` and that item's kind fits the slot (hand weapons in slots 10, 11 and 12, a ranged weapon in slot 13, a shield in slot 11). A piece that resolves to a kind that does not fit SHALL be skipped.

#### Scenario: Matching weapon
- **WHEN** an NPC's right-hand piece resolves to a mace model
- **THEN** that mace is drawn on the right-hand socket and follows the NPC's animation

#### Scenario: Wrong kind for the slot
- **WHEN** a left-hand piece resolves to a bow
- **THEN** nothing is drawn for that slot

#### Scenario: Unknown number
- **WHEN** a piece's number is absent from `items.csv`
- **THEN** nothing is drawn for that slot and the skipped count rises

### Requirement: Each item model is converted once and shared
Held-item models SHALL be converted to one shared copy per item under the converted data folder and loaded on first use, not copied into each creature that carries one.

#### Scenario: Two NPCs, one mace
- **WHEN** two NPCs carry the same mace
- **THEN** the client loads that model once

### Requirement: Armour is exact only when a real mapping is found
The change SHALL record the outcome of a bounded search for the install's armour mapping. If a mapping is found, the client SHALL resolve armour pieces exactly from it. If none is found, armour SHALL be drawn by tier, and the README SHALL say that armour appearance is approximate.

#### Scenario: Mapping found
- **WHEN** the search identifies the table the client uses
- **THEN** armour pieces resolve through it and the fallback is used only for numbers it lacks

#### Scenario: No mapping found
- **WHEN** the search ends without a mapping
- **THEN** the outcome and where it looked are written in the design, and armour is drawn by tier

### Requirement: Fallback armour is chosen by the armour type the database gives each piece
In the fallback, the client SHALL classify each worn armour number by the armour type and realm of the imported items that use it, dress the matching body slot in the tier mesh for that type in a texture family of that realm, and tint it with the piece's `Color` from `tints.csv`.

#### Scenario: Full armour set
- **WHEN** an NPC wears torso, legs, arms, hands and feet pieces whose numbers the database gives as Albion plate
- **THEN** each slot shows its plate tier in Albion plate textures, tinted by its colour

#### Scenario: Type the realm lacks
- **WHEN** a piece is plate and its realm is Midgard
- **THEN** it is dressed as Midgard chain, the nearest lighter type that realm has

#### Scenario: Torso only
- **WHEN** an NPC wears only a torso piece
- **THEN** only the torso slot changes and the other slots keep the base model's

#### Scenario: Unclassified number
- **WHEN** a worn armour number is used by no item in the database
- **THEN** that slot keeps the base model and the skipped count rises

#### Scenario: Unknown colour
- **WHEN** a piece's colour id is not in `tints.csv`
- **THEN** it is drawn with its texture's own colours
