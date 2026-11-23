## ADDED Requirements

### Requirement: Keep doors
A keep door SHALL have health and armour by the keep's level. Enemy players SHALL be able to attack it and destroy it, which opens the way through. The owning realm SHALL pass through a standing door by using it. A destroyed door SHALL return with full health when the keep changes hands or after its repair time.

#### Scenario: Breaking a door
- **WHEN** enemy players bring a door's health to zero
- **THEN** the door opens permanently until repaired, and the zone is told the door is down

#### Scenario: Owners pass
- **WHEN** an owning-realm player uses a standing door
- **THEN** they pass to the other side

### Requirement: Guards and lord
A keep SHALL spawn guards and a lord of its owner realm from its data. Guards SHALL attack enemy players in range and help each other.

#### Scenario: Guards defend
- **WHEN** an enemy player comes within a guard archer's range
- **THEN** the guard archer shoots at them

### Requirement: Ownership
A keep SHALL be owned by one realm. Killing its lord SHALL transfer it to the killer's realm.

#### Scenario: Owner shown
- **WHEN** a player opens the Keep window for a keep
- **THEN** it shows the owner realm and the door states
