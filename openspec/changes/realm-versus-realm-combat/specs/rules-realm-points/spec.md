## ADDED Requirements

### Requirement: Realm point value of a kill
A player killed by an enemy realm in an RvR area SHALL be worth realm points by their level and realm rank, under the rule book's formula, citing its source and status.

#### Scenario: Level 50 victim
- **WHEN** a level 50 RR2 player is killed
- **THEN** the kill is worth the rule book's figure for level 50 at RR2

### Requirement: Sharing
The value SHALL be split among enemy players who damaged the victim, by their share of the damage. A share earned by a grouped player SHALL be divided among the group members in range.

#### Scenario: Two attackers
- **WHEN** two solo attackers did 75 and 25 percent of the damage
- **THEN** they receive 75 and 25 percent of the value

### Requirement: Repeat kills are worth less
A victim killed again soon after their last death SHALL be worth less, by the rule book's reduction.

#### Scenario: Killed twice in a minute
- **WHEN** a player is killed twice within the reduction window
- **THEN** the second kill is worth less than the first

### Requirement: Realm rank
Realm rank and level SHALL follow the published realm rank table from total realm points, up to the 1.65 cap. Each realm level SHALL give one realm skill point. Each full rank SHALL add +1 to all specialization skills.

#### Scenario: Reaching RR2
- **WHEN** a player's realm points cross the RR2L0 threshold
- **THEN** they read the rank message, gain the skill points due, and their skills rise by 1

### Requirement: Registered and tested
Each rule SHALL have table tests per scenario and a registry entry.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs
- **THEN** the realm point rules are listed with their status
