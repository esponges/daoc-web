## ADDED Requirements

### Requirement: Berserk
Berserk SHALL be a timed stance of the Berserker that raises melee damage and lowers the user's own defence for its duration, with a reuse timer. Its amounts per ability level SHALL be in the rule book with sources and a status.

#### Scenario: Berserk raises damage
- **WHEN** a Berserker with Berserk active hits the same target with the same weapon as without it
- **THEN** the base damage is higher by the rule book's factor, and the Berserker is easier to hit

#### Scenario: Reuse
- **WHEN** a Berserker uses Berserk again before its reuse timer ends
- **THEN** it is refused with the remaining time

### Requirement: Charge
Charge SHALL raise the user's run speed for a short time even in combat, with a reuse timer. Slows and snares SHALL not apply while it lasts.

#### Scenario: Charge in combat
- **WHEN** a Savage in combat uses Charge
- **THEN** their speed rises for its duration and a snare cast on them during it has no effect

### Requirement: Enhanced Evade
Enhanced Evade SHALL raise the evade chance from the melee rules by its rule-book amount, with the evade cap still applying.

#### Scenario: Evade raised
- **WHEN** a Berserker with Enhanced Evade is attacked from the front
- **THEN** the evade chance is higher than with plain Evade at the same Dexterity and Quickness

### Requirement: QuickCast
QuickCast SHALL make the user's next spell cast with no cast time, with a reuse timer, and SHALL cost more power than the spell normally does, by its rule-book factor.

#### Scenario: Instant next spell
- **WHEN** a Runemaster uses QuickCast and then casts a bolt
- **THEN** the bolt fires without a cast bar, at the higher power cost, and the next spell after it has its normal cast time

### Requirement: Safe Fall is counted
Safe Fall SHALL be reported by the parity census as implemented, because the falling rules already reduce damage by Stealth specialization.

#### Scenario: Census
- **WHEN** `cmd/parity` runs
- **THEN** Safe Fall is listed as implemented

### Requirement: Danger Sense
Danger Sense SHALL warn its holder when a hidden enemy-realm player comes within its range.

#### Scenario: Warning
- **WHEN** a hidden enemy player comes within range of a Shadowblade with Danger Sense
- **THEN** the Shadowblade reads "You sense danger nearby!" once per approach

### Requirement: Distraction
Distraction SHALL be aimed at a ground target. Creatures in its radius that are not in combat SHALL turn to face the point.

#### Scenario: Creatures look away
- **WHEN** a Shadowblade uses Distraction on ground near two idle creatures
- **THEN** both turn to face the ground target and do not see the Shadowblade's approach any sooner than usual

### Requirement: Registered and tested
Each rule in this change SHALL have a table test per numbered scenario and a registry entry with its sources and status, so `cmd/parity` counts it.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs after this change
- **THEN** abilities read 28 or more of 30, and each remaining gap has a reason
