## ADDED Requirements

### Requirement: The power pool
A caster's maximum power SHALL be set by their level, their class's casting stat and the stat's modifier from the source named in the spec's Sources section. A class with no spells SHALL have no power. Status: provisional until a source with the formula is adopted.

#### Scenario: A non-caster
- **WHEN** a Warrior's sheet is built
- **THEN** it shows no power

#### Scenario: Levelling
- **WHEN** a Healer gains a level
- **THEN** their maximum power rises per the formula and is shown on the power bar

### Requirement: Power regeneration
Power SHALL regenerate on a fixed tick, faster out of combat than in it, and SHALL not regenerate while the caster is dead. Casting SHALL spend the spell's power cost when the cast completes.

#### Scenario: Resting
- **WHEN** a Runemaster with half their power sits out of combat for a minute
- **THEN** their power is higher by the out-of-combat rate times the ticks passed

#### Scenario: Interrupted cast
- **WHEN** a cast is interrupted
- **THEN** no power is spent

### Requirement: Registered and tested
Each rule in this spec SHALL have a table test in `internal/server/rules` for each numbered scenario and a registry entry so that `cmd/parity` counts it, with its sources and status listed in the spec.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs
- **THEN** the power rules are listed with their status
