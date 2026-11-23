## ADDED Requirements

### Requirement: The movement step moves vertically
The shared movement step SHALL carry vertical speed: a jump input on the ground starts a jump, a body in the air falls under gravity until it meets the ground, and a body in water is held at the surface. The Go and JS steps SHALL stay one definition, checked against shared vectors that include jumps, ledges and water.

#### Scenario: Shared vectors
- **WHEN** the Go tests write the movement vectors and the JS test replays them
- **THEN** both agree on every step, including the steps of a jump and of a fall off a ledge

#### Scenario: The jump flag
- **WHEN** a client holds the jump key
- **THEN** its input runs carry the jump flag, and a server built before it ignores the flag as an unknown bit

### Requirement: Posture on the entity
A player entity SHALL carry its posture (`ps`: seated, in the air or swimming; absent when standing on the ground) and, in the air, its vertical speed (`vz`), as additive fields.

#### Scenario: A seated player in a snapshot
- **WHEN** a seated player is in another player's snapshot
- **THEN** their entity has `ps` "sit"
