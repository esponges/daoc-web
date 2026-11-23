# rules-falling Specification

## Purpose
The rule book's entry for falling: when a drop hurts a player, how much, and how the Safe Fall ability of the Stealth line softens it.

## Requirements

### Requirement: Falling damage
A landing SHALL cost nothing when the fall is no higher than the safe height, 480 units. Above it the loss SHALL be (height − 480) / 1920 of maximum health, at most all of it. Status: documented that falls hurt since patch 1.38; the numbers are this project's and provisional.

#### Scenario: A short drop
- **WHEN** a player lands from a fall below the safe height
- **THEN** they take no damage and no message

#### Scenario: A long drop
- **WHEN** a player lands from a fall twice the safe height
- **THEN** they lose (960 − 480) / 1920 = 25 percent of maximum health, and read "You take N damage from the fall!"

#### Scenario: A fall can kill
- **WHEN** a player at low health lands from a fall that costs more than they have
- **THEN** they die, and lose no experience, since only a creature's kill costs it

### Requirement: Safe Fall
Safe Fall SHALL be one level for each ten points of Stealth specialization (at most 5). Each level SHALL raise the safe height by 240 units and take 15 percent off the damage. Status: documented in patch 1.38 for the shape and the one level per ten points; the amounts per level are provisional.

#### Scenario: A Shadowblade with Stealth 30
- **WHEN** a Shadowblade with Safe Fall 3 and a Warrior land from the same high fall
- **THEN** the Shadowblade takes less damage than the Warrior

### Requirement: Swimming and water landings
A fall that ends in water SHALL do no damage.

#### Scenario: Into a lake
- **WHEN** a player falls from a cliff into deep water
- **THEN** they take no damage and swim

### Requirement: Registered and tested
Each rule in this spec SHALL have a table test in `internal/server/rules` for each numbered scenario and a registry entry so that `cmd/parity` counts it, with its sources and status.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs
- **THEN** the falling rules are listed with their status
