## ADDED Requirements

### Requirement: Temple gates
A relic temple's gate SHALL be closed to enemies while its realm holds the keeps that guard it. It SHALL open when the 1.65 rule says those keeps have fallen.

#### Scenario: Gate opens
- **WHEN** Midgard loses the keeps guarding its Strength temple
- **THEN** the temple gate opens to enemies and every realm is told

### Requirement: Carrying a relic
An enemy player SHALL be able to pick up a relic from its pad. The carrier SHALL not benefit from speed effects or use teleports, and SHALL be shown on the war map. Dying SHALL drop the relic at the spot. A relic left on the ground for the rule book's time SHALL return to its home temple.

#### Scenario: Speed refused
- **WHEN** a relic carrier receives a speed song
- **THEN** their speed is unchanged

#### Scenario: Dropped relic returns
- **WHEN** a dropped relic lies untouched for the return time
- **THEN** it returns to its home pad and every realm is told

### Requirement: Capture
A carrier who places an enemy relic on an empty pad of the same type in their own realm's temple SHALL capture it for their realm.

#### Scenario: Captured
- **WHEN** a Midgard carrier places Albion's Strength relic on Midgard's Strength temple pad
- **THEN** Midgard holds it and every realm is told

### Requirement: Relic bonuses
Each enemy Strength relic a realm holds SHALL raise its players' melee damage by the rule book's percentage. Each enemy Power relic SHALL raise magic damage likewise.

#### Scenario: One Strength relic
- **WHEN** Midgard holds one enemy Strength relic
- **THEN** a Midgard player's melee damage is higher by one relic's percentage, and the bonuses list shows it
