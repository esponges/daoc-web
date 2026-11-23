# rules-resting Specification

## Purpose
The rule book's entry for resting: how much faster a seated player regains health, power and endurance than one standing out of combat.

## Requirements

### Requirement: Seated regeneration
A seated player out of combat SHALL regain health, power and endurance at twice the standing amount per tick, the ticks being those of `rules-character-stats` and `rules-power`. Status: provisional. The 1.87 notes, as quoted on the Dawn of Light forum, doubled standing ticks to match sitting; no 1.65 page gives the factor itself.

#### Scenario: Level 20 seated
- **WHEN** a level 20 Warrior below full health sits out of combat for 30 seconds
- **THEN** they regain 140 hit points, where standing they regain 70

#### Scenario: Seated caster
- **WHEN** a level 50 Runemaster sits out of combat for one power tick
- **THEN** they regain twice 5 + 50 / 2.75 power, about 46.4

#### Scenario: Seated endurance
- **WHEN** a Warrior with 40 endurance sits out of combat for one second
- **THEN** they have 52 endurance

### Requirement: Registered and tested
Each rule in this spec SHALL have a table test in `internal/server/rules` for each numbered scenario and a registry entry so that `cmd/parity` counts it, with its sources and status.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs
- **THEN** the resting rule is listed with its status
