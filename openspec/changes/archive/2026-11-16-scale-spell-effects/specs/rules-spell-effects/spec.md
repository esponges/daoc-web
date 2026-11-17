## MODIFIED Requirements

### Requirement: Damage over time ticks
A damage-over-time spell SHALL hit every `Frequency` tenths of a second for `Duration` seconds. A creature's spell SHALL hit for its `Damage` after the resist, with no variance. A player's spell from their book SHALL hit, on each tick, for its `Damage` × (casting stat + 200) / 275 × a roll uniform between a low end and 1.13, then the resist, then at most three times the spell's `Damage`, rounded down. The low end is 0.13 + (the caster's spec in the spell's line − 1) / the target's level, between 0 and 1.13. There is no level-difference term. The casting stat and spec are those of the caster when the spell landed. A second cast of the same spell on the same target SHALL restart its duration and take the new cast's stat and spec, not stack. Status: reference (DOLSharp states the range; OpenDAoC-Core uses the general spell variance for it, and the DOLSharp range is followed).

#### Scenario: Five ticks
- **WHEN** a creature's spell with Damage 10, Frequency 40 and Duration 20 lands on a target with no resist
- **THEN** the target takes 10 five times, 4 seconds apart

#### Scenario: Recast
- **WHEN** the same spell lands again on that target after two ticks
- **THEN** one effect remains, with 20 seconds to run

#### Scenario: A player's tick at spec
- **WHEN** a player with 100 in their casting stat and a spec of 20 casts a 10-damage spell at a level 20 target and the tick rolls the middle of the range
- **THEN** the range is 1.08 to 1.13 and the tick does 12

#### Scenario: A low spec tick
- **WHEN** the same caster has a spec of 1 and the tick rolls the worst result
- **THEN** the low end is 0.13 and the tick does 1

#### Scenario: The tick cap
- **WHEN** a caster with 600 in their casting stat casts a 5-damage spell and the tick rolls the best result
- **THEN** the tick does at most 15

#### Scenario: A player's tick uses the cast
- **WHEN** a player's damage-over-time spell has landed and the caster's casting stat then falls
- **THEN** the ticks keep the stat the caster had when it landed

## ADDED Requirements

### Requirement: A damage add is a damage per second
A damage-add spell's `Damage` SHALL be a damage per second. On each landed blow it SHALL add the lesser of `Damage` and the cap for the caster's level, × the blow's swing time in seconds × a roll uniform between 0.5 and 1, rounded down, after the target's resist. The cap is (1.2 + 0.3 × level) × 0.7 damage per second. A damage add whose `Damage` is negative SHALL add that percent of the blow's own damage. It does not depend on the caster's stat or spec. The strongest damage add on a player counts. Status: reference (DOLSharp states the formula and the cap; OpenDAoC-Core scales by swing time with the general variance; the DOLSharp roll is followed).

#### Scenario: A slow weapon
- **WHEN** a level 20 caster's 4.5 per second damage add lands on a 3.0 second swing and rolls the middle
- **THEN** it adds 10

#### Scenario: The damage add cap
- **WHEN** a level 3 caster's 9 per second damage add lands on a 3.0 second swing and rolls the best result
- **THEN** it adds 4, the cap 1.47 times 3 seconds

#### Scenario: A percent add
- **WHEN** a damage add of −20 lands on a blow of 50 damage
- **THEN** it adds 10

#### Scenario: The strongest counts
- **WHEN** two damage-add spells are on a player
- **THEN** each blow adds from the stronger one only
