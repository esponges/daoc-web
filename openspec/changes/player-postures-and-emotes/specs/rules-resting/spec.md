# Spec Delta

## Purpose

The rule book's entry for resting: how much faster a seated player regains health, power and endurance than one standing out of combat.

## ADDED Requirements

### Requirement: Seated regeneration
A seated player out of combat SHALL regain health, power and endurance at twice the out-of-combat amount per tick. The ticks stay as `rules-character-stats` and `rules-power` give them. Status: provisional. The later patch that doubled standing regeneration "to match the bonus while sitting" implies a factor of two at 1.65. Sources: the 1.87 patch notes as quoted on the Dawn of Light forum (standing ticks doubled to match sitting, so sitting was twice the standing bonus before it). No 1.65-era page gives the number itself, so the status stays provisional.

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
