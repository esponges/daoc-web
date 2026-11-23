## ADDED Requirements

### Requirement: Entering the frontier
Each realm SHALL have a portal from its home zones to its border keep in the frontier, and back. The frontier zones SHALL be RvR areas.

#### Scenario: Into the frontier
- **WHEN** a Midgard player uses the frontier portal NPC
- **THEN** they arrive at Midgard's border keep in the frontier

### Requirement: Keeps at their owners
Every frontier keep and tower SHALL be in the world with its doors, guards and lord, owned by its data's realm until captured.

#### Scenario: Census
- **WHEN** `cmd/parity` runs
- **THEN** keeps and keep doors read as implemented out of their totals
