# rules-spell-effects Specification

## Purpose
Defines, for the spell types creatures cast in 1.65 Midgard, whether a spell lands, what it does to its target, how resists reduce it, and how long it lasts, using each spell's imported values.

## Requirements

### Requirement: A hostile spell lands by a hit roll
A hostile spell SHALL land with chance 85% + (caster level − target level) / 2 %, clamped to 50%–99%. A spell that does not land SHALL do nothing, and both sides SHALL be told the target resisted it. Status: reference, until a 1.65 source is recorded.

#### Scenario: Even levels
- **WHEN** a level 10 creature casts at a level 10 Warrior
- **THEN** the chance to land is 85%

#### Scenario: A higher caster
- **WHEN** a level 16 creature casts at a level 10 Warrior
- **THEN** the chance is 88%

#### Scenario: Far below
- **WHEN** a level 1 creature casts at a level 90 target
- **THEN** the chance is clamped to 50%

### Requirement: Resists reduce spell damage
Spell damage SHALL be multiplied by (1 − resist / 100), where `resist` is the target's resist to the spell's damage type from race and items, capped at 70%. Status: reference for the cap's value, until a 1.65 source is recorded.

#### Scenario: A resisted fraction
- **WHEN** a spell does 100 before resists to a target with 10% resist to its type
- **THEN** it does 90

### Requirement: Direct damage and bolts vary
A creature's direct-damage or bolt spell SHALL do its `Damage` × a variance uniform between 0.75 and 1.25, then the resist, rounded down; a spell of the no-variance type SHALL use a variance of 1. A player's direct-damage or bolt spell from their book SHALL do its `Damage` × (casting stat + 200) / 275, × a roll uniform between a low end and a high end of 1.25, then the resist, then at most three times the spell's `Damage`, rounded down. The low end is 0.25 + (the caster's spec in the spell's line − 1) / the target's level. A spec above the target's level adds 0.005 per level over, and a caster above the target's level adds 0.02 per level of difference, to both ends; the high end is at least 0.25 and the low end between 0 and the high end. A no-variance spell from a player SHALL do its `Damage` scaled by neither. Status: reference (DOLSharp and OpenDAoC-Core; Mythic's Grab Bag answers confirm that a spec at or above the spell's level gives full effect and that damage starts at +25 percent, with no figures).

#### Scenario: Middle of the range
- **WHEN** a creature's 40-damage direct-damage spell rolls a variance of 1.0 against 10% resist
- **THEN** it does 36

#### Scenario: No variance
- **WHEN** a 40-damage no-variance spell lands on a target with no resist
- **THEN** it does exactly 40

#### Scenario: A fully specced caster
- **WHEN** a level 20 caster with 100 in their casting stat and a spec of 20 casts a 40-damage spell at a level 20 target, rolling the middle of the range
- **THEN** the range is 1.2 to 1.25 and the spell does 53

#### Scenario: A low spec
- **WHEN** the same caster has a spec of 5 and rolls the worst result
- **THEN** the low end is 0.45 and the spell does 19

#### Scenario: The cap
- **WHEN** a level 50 caster with 300 in their casting stat and a spec of 50 casts a 10-damage spell at a level 1 target
- **THEN** the range is far above the cap and the spell does 30

### Requirement: A bolt can be blocked
A bolt that lands SHALL be tested against the target's block chance, as a melee blow from the front would be. A blocked bolt SHALL do no damage. Status: reference.

#### Scenario: Blocked bolt
- **WHEN** a bolt reaches a Warrior with a shield facing the caster and the block roll succeeds
- **THEN** the bolt does no damage and both sides see it blocked

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

### Requirement: Slows lower speed
A speed-decrease spell SHALL multiply the target's movement speed by (1 − `Value` / 100) for `Duration` seconds. The strongest active slow SHALL apply. Status: reference.

#### Scenario: A 40% slow
- **WHEN** a 40% slow lands on a running Warrior
- **THEN** their speed is 60% of normal for its duration, and the client's predicted movement agrees

### Requirement: Mesmerize holds until hurt
A mesmerize spell SHALL stop its target moving, swinging, casting and using styles for `Duration` seconds. Any damage to the target SHALL end it. Status: reference.

#### Scenario: Broken by damage
- **WHEN** a mesmerized player is hit by another creature
- **THEN** the mesmerize ends at once and the player can act

### Requirement: Stun holds through damage
A stun spell SHALL stop its target moving, swinging, casting and using styles for `Duration` seconds. Damage SHALL NOT end it. Status: reference.

#### Scenario: Hit while stunned
- **WHEN** a stunned player is hit
- **THEN** the stun continues to its end

### Requirement: Life drain heals its caster
A life-drain spell SHALL do damage as direct damage does, and SHALL heal its caster by that damage × `LifeDrainReturn` / 100, up to the caster's maximum health. `LifeDrainReturn` is a percent: the data's 117 drain spells use values from 0 to 100. Status: reference.

#### Scenario: Drain
- **WHEN** a drain with Damage 40, return 50, rolls a variance of 1.0 on a target with no resist
- **THEN** the target loses 40 and the caster regains 20

### Requirement: Heals restore health
A creature's heal spell SHALL restore `Value` health to its target, up to its maximum. A player's heal from their book SHALL restore a roll between a least and a most, up to the target's maximum: the most is 1.25 × `Value` and the least is `Value` × (0.25 + (the caster's spec in the spell's line − 1) / the spell's level), at most 1.25, and at least 1. A heal whose value is negative SHALL restore that percent of the caster's maximum health. Status: reference (DOLSharp and OpenDAoC-Core; Mythic's Grab Bag answers confirm that heal variance narrows as spec nears the spell's level, with no figures).

#### Scenario: A creature heals itself
- **WHEN** a creature at 100 of 300 health casts a 120-point heal on itself
- **THEN** it has 220

#### Scenario: A player heal at spec
- **WHEN** a caster with a spec of 20 casts a level 20 heal of 120
- **THEN** it restores between 144 and 150

#### Scenario: A player heal below spec
- **WHEN** a caster with a spec of 5 casts a level 20 heal of 120
- **THEN** it restores between 54 and 150

#### Scenario: A percent heal
- **WHEN** a caster with 720 health casts a heal whose value is −40
- **THEN** it restores 288

### Requirement: Effects show and end
Each lasting effect SHALL be sent to its target with its spell name, icon id and time left, and SHALL be removed when it ends, when its caster's spell is replaced, or when its target dies. Effects SHALL NOT survive a logout. Status: provisional for the logout rule.

#### Scenario: Effect icons
- **WHEN** a slow lands on a player
- **THEN** their client shows an effect with the spell name and a countdown, gone when the slow ends

### Requirement: A spell rule's status follows its source
A spell rule SHALL NOT be marked above provisional unless its spec names a source for the formula, and SHALL NOT be marked documented unless that source is a published account of 1.65 itself. Where two sources disagree, the spec SHALL say so and name the one followed.

#### Scenario: A reference rule names its source
- **WHEN** a spell rule is marked reference
- **THEN** its spec names the community server or table the formula was understood from, in the project's own words

#### Scenario: No source found
- **WHEN** no source for a formula can be found
- **THEN** the rule stays provisional and `findings.md` records the search

### Requirement: Spell numbers agree with the imported spell table
For each imported spell of an implemented type, the amount the server would deal or restore for a reference caster and target SHALL be derivable from the spell's own `Damage` or `Value`, and a spell whose amount differs from that derivation SHALL be reported.

#### Scenario: A direct damage spell
- **WHEN** the check runs over a 40-damage direct-damage spell at the middle of its range, with no resist
- **THEN** the amount is within the range the rule gives for that spell's base and the spell is not reported

#### Scenario: A heal
- **WHEN** the check runs over a 120-value heal
- **THEN** the amount is within 0.25 to 1.25 times 120, up to the target's maximum

#### Scenario: A mismatch
- **WHEN** a spell's amount falls outside that range
- **THEN** the check names the spell id and both amounts

#### Scenario: No database
- **WHEN** `server/data/community` is absent
- **THEN** the check is skipped and the rest of the tests run

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
