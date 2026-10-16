## MODIFIED Requirements

### Requirement: Worn gear is drawn on NPCs
The client SHALL draw each piece of an NPC's equipment that it can resolve to art from the user's install, following the resolution rules of `npc-gear-mapping`: held items from the install's item table when the kind fits the slot, and armour exactly when a real mapping exists or by tier otherwise, tinted by the piece's colour.

#### Scenario: Armoured NPC
- **WHEN** an NPC with torso, leg and boot pieces comes into view
- **THEN** those slots show armour resolved under `npc-gear-mapping` instead of the base model's default

#### Scenario: Armed NPC
- **WHEN** an NPC has a right-hand or two-hand weapon that resolves
- **THEN** the weapon is drawn in its hand and follows its animation
