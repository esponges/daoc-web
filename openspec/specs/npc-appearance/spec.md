# npc-appearance Specification

## Purpose
How an NPC is drawn from the equipment the imported data gives it: worn and held gear on its model, a quiet fallback when gear cannot be resolved, and heads and faces only as far as the data supports.

## Requirements

### Requirement: Worn gear is drawn on NPCs
The client SHALL draw each piece of an NPC's equipment that it can resolve to art from the user's install, following the resolution rules of `npc-gear-mapping`: held items from the install's item table when the kind fits the slot, and armour exactly when a real mapping exists or by tier otherwise, tinted by the piece's colour.

#### Scenario: Armoured NPC
- **WHEN** an NPC with torso, leg and boot pieces comes into view
- **THEN** those slots show armour resolved under `npc-gear-mapping` instead of the base model's default

#### Scenario: Armed NPC
- **WHEN** an NPC has a right-hand or two-hand weapon that resolves
- **THEN** the weapon is drawn in its hand and follows its animation

### Requirement: Unresolved gear degrades quietly
A piece whose model number cannot be resolved SHALL be skipped without an error, leaving the base model for that slot, and the client SHALL count skipped pieces for the developer panel. With no converted equipment data present the client SHALL draw NPCs exactly as it does today.

#### Scenario: Unknown model number
- **WHEN** an NPC wears a piece the lookup does not know
- **THEN** the rest of its gear still draws and nothing is logged to the player

#### Scenario: Data folder absent
- **WHEN** the converted equipment lookup is missing
- **THEN** NPCs draw as plain models and the game works

### Requirement: Faces and heads follow what the data supports
The change SHALL record whether NPC faces and heads are selected by equipment data or belong to the base model. The client SHALL draw a face only when that selection is known.

#### Scenario: Head not selectable
- **WHEN** the investigation finds no equipment-driven head selection
- **THEN** the finding is written in the design and NPCs keep the base model's head
