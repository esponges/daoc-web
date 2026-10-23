# rules-power Specification

## Purpose
TBD - created by archiving change next-midgard-classes-and-spells. Update Purpose after archive.

## Requirements

### Requirement: The power pool
A caster's maximum power SHALL be five times their level plus their class's casting stat above 50, and at least 5. A class with no spells SHALL have no power. Status: reference (DOLSharp and OpenDAoC-Core agree; no published 1.65 account gives the formula, only that the pool depends on the casting stat).

#### Scenario: A non-caster
- **WHEN** a Warrior's sheet is built
- **THEN** it shows no power

#### Scenario: Levelling
- **WHEN** a Healer gains a level
- **THEN** their maximum power rises by five, and a Healer with 70 piety has 25 at level 1 and 30 at level 2

### Requirement: Power regeneration
Power SHALL regenerate on a fixed tick of three seconds, by five plus the caster's level divided by 2.75, in combat or out of it, and SHALL not regenerate while the caster is dead. Casting SHALL spend the spell's power cost when the cast completes. Status: provisional for combat: the community servers state this rate for the time after patch 1.87, which made combat ticks twice as effective, so 1.65 combat regeneration was lower by an amount no source gives.

#### Scenario: Resting
- **WHEN** a level 50 Runemaster sits for a minute
- **THEN** their power is higher by about 23 times the ticks passed

#### Scenario: Interrupted cast
- **WHEN** a cast is interrupted
- **THEN** no power is spent

### Requirement: Registered and tested
Each rule in this spec SHALL have a table test in `internal/server/rules` for each numbered scenario and a registry entry so that `cmd/parity` counts it, with its sources and status listed in the spec.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs
- **THEN** the power rules are listed with their status

### Requirement: A spell's cost is read from its power
A spell whose imported power is positive SHALL cost that much power. A spell whose imported power is negative SHALL cost that percent of the caster's pool, at least 1. The focus bonus that lowers a cost in the community servers is not modelled. Status: reference.

#### Scenario: A fixed cost
- **WHEN** a spell with power 3 is cast by a caster with a pool of 90
- **THEN** it costs 3

#### Scenario: A share of the pool
- **WHEN** a spell with power −10 is cast by a caster with a pool of 90
- **THEN** it costs 9
