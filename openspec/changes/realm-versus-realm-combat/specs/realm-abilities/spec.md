## ADDED Requirements

### Requirement: Buying realm abilities
A player SHALL buy realm abilities of their class from the 1.65 set at a trainer's Realm tab, with realm skill points, at each ability's cost per level.

#### Scenario: Buying
- **WHEN** a Warrior with 5 realm skill points buys Augmented Strength level 1
- **THEN** the points are spent and their Strength rises by the ability's amount

### Requirement: Active abilities
An active realm ability SHALL take effect when used and start its reuse timer. Using it again before the timer ends SHALL be refused with the time left.

#### Scenario: Purge
- **WHEN** a mesmerized player with Purge uses it
- **THEN** the mesmerize and other negative effects end and Purge's timer starts

### Requirement: Respec
A realm respec SHALL refund all spent realm skill points.

#### Scenario: Respec
- **WHEN** a player uses a realm respec
- **THEN** all bought realm abilities are removed and their points returned

### Requirement: Only 1.65 abilities
The realm ability set SHALL be the 1.65 set per class. Later abilities in the community data SHALL be excluded and counted as not classic.

#### Scenario: Census
- **WHEN** `cmd/parity` runs
- **THEN** a realm abilities row lists implemented of the 1.65 total, and the excluded ones with reasons
