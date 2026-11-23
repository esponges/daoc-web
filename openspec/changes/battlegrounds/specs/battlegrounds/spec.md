## ADDED Requirements

### Requirement: Entry by level and rank
A battleground's portal SHALL admit a player whose level is within its range and whose realm rank is at or below its cap, sending them to their realm's portal keep. Others SHALL be refused with the reason.

#### Scenario: Thidranki entry
- **WHEN** a level 22 Midgard player uses the Thidranki portal
- **THEN** they arrive at Midgard's portal keep in Thidranki

#### Scenario: Too high
- **WHEN** a level 26 player uses the Thidranki portal
- **THEN** it is refused, naming the level range

### Requirement: Outgrowing a battleground
A player inside a battleground who passes its level or rank cap SHALL be warned, and SHALL be moved to their realm's border keep after a grace time.

#### Scenario: Levelling out
- **WHEN** a player in Thidranki reaches level 25
- **THEN** they read a warning and are moved out after the grace time

### Requirement: Central keep capture
Killing the central keep's lord SHALL give the keep to the killer's realm, announce it to every player in the battleground, and respawn the guards of the new owner.

#### Scenario: Capture
- **WHEN** a Midgard group kills the lord of an Albion-held central keep
- **THEN** the keep turns to Midgard and the capture is announced
