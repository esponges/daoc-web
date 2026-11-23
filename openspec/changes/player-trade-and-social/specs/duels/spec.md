## ADDED Requirements

### Requirement: Challenge and answer
A player SHALL be able to challenge a living player of the same realm in range with /duel. The target SHALL accept or decline within 30 seconds, and silence counts as declining.

#### Scenario: Accepted
- **WHEN** Thora challenges Bjorn and Bjorn accepts
- **THEN** both read that the duel has begun and can attack each other

### Requirement: Duel rules
During a duel only the two duellists SHALL be able to harm each other. A duel SHALL end when one would fall to zero health, leaving them at 1 health as the loser. It SHALL also end, with no winner, when they are more than 2500 units apart, one changes zone, or a third party attacks either. A duel SHALL give no experience and cause no death penalty.

#### Scenario: Defeat
- **WHEN** Bjorn's health would reach zero from Thora's blow
- **THEN** Bjorn is left at 1 health, both stop fighting, and the zone reads "Thora has defeated Bjorn in a duel!"

#### Scenario: Interference
- **WHEN** a creature attacks Thora during the duel
- **THEN** the duel ends with no winner

### Requirement: Player-versus-player combat
Attacks between duellists SHALL use the melee, style and spell rules with the defender's own defences (evade, parry, block, armour factor and absorption, resists) in place of a creature's.

#### Scenario: Shield block
- **WHEN** a Warrior with a shield is swung at by a duellist from the front
- **THEN** the block chance from the melee rules applies
