## ADDED Requirements

### Requirement: Berserk
Berserk SHALL be a timed stance of the Berserker that raises melee damage and lowers the user's own evade, parry and block for its duration, with a reuse timer. Its amounts per ability level SHALL be in the rule book with sources and a status. No source found gives them (a later developer post says Berserk made a Berserker's blows critical, which this project has not adopted), so they are provisional.

#### Scenario: Berserk raises damage
- **WHEN** a Berserker with Berserk active hits the same target with the same weapon as without it
- **THEN** the base damage is higher by the rule book's factor, and the Berserker's evade, parry and block chances are lower by the rule book's share

#### Scenario: Reuse
- **WHEN** a Berserker uses Berserk again before its reuse timer ends
- **THEN** it is refused with the remaining time

### Requirement: Charge
Charge SHALL raise the user's run speed for a short time even in combat, with a reuse timer. Slows and snares SHALL not apply while it lasts.

#### Scenario: Charge in combat
- **WHEN** a Savage in combat uses Charge
- **THEN** their speed rises for its duration and a snare cast on them during it has no effect

### Requirement: QuickCast
QuickCast SHALL make the user's next spell cast with no cast time, with a reuse timer. It SHALL end when that spell goes off or after its window.

#### Scenario: Instant next spell
- **WHEN** a Runemaster uses QuickCast and then casts a bolt
- **THEN** the bolt fires without a cast bar, and the next spell after it has its normal cast time

### Requirement: Enhanced Evade is left out, with its reason
Enhanced Evade SHALL stay unimplemented while no source says what it did or that 1.65 had it. The parity census SHALL list it with that reason and the date searched.

#### Scenario: Census
- **WHEN** `cmd/parity` runs
- **THEN** Enhanced Evade is listed as a gap with its reason

### Requirement: Safe Fall is counted
Safe Fall SHALL be reported by the parity census as implemented, because the falling rules already reduce damage by Stealth specialization.

#### Scenario: Census Safe Fall
- **WHEN** `cmd/parity` runs
- **THEN** Safe Fall is listed as implemented

### Requirement: Danger Sense
Danger Sense SHALL warn its holder when a hidden enemy-realm player comes within its range. It is inert until a second realm can be in a zone with the player.

#### Scenario: Warning
- **WHEN** a hidden enemy player comes within range of a Shadowblade with Danger Sense
- **THEN** the Shadowblade reads "You sense danger nearby!" once per approach

### Requirement: Distraction
Distraction SHALL be aimed at a ground target. Idle creatures in its radius SHALL turn to face the point.

#### Scenario: Creatures look away
- **WHEN** a Shadowblade uses Distraction on ground near two idle creatures
- **THEN** both turn to face the ground target

### Requirement: Registered and tested
Each rule in this change SHALL have a table test per numbered scenario and a registry entry with its sources and status, so `cmd/parity` counts it.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs after this change
- **THEN** abilities read 27 of 30, and each remaining gap (Climb Walls, Enhanced Evade, Call of a thousand storms) has a reason
