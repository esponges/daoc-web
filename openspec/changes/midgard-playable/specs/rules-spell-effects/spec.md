# Spec Delta

## Purpose

Defines, for the spell types creatures cast in 1.65 Midgard, whether a spell lands, what it does to its target, how resists reduce it, and how long it lasts, using each spell's imported values.

## ADDED Requirements

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
A direct-damage or bolt spell SHALL do its `Damage` × a variance uniform between 0.75 and 1.25, then the resist, rounded down. A spell of the no-variance type SHALL use a variance of 1. Status: provisional.

#### Scenario: Middle of the range
- **WHEN** a 40-damage direct-damage spell rolls a variance of 1.0 against 10% resist
- **THEN** it does 36

#### Scenario: No variance
- **WHEN** a 40-damage no-variance spell lands on a target with no resist
- **THEN** it does exactly 40

### Requirement: A bolt can be blocked
A bolt that lands SHALL be tested against the target's block chance, as a melee blow from the front would be. A blocked bolt SHALL do no damage. Status: reference.

#### Scenario: Blocked bolt
- **WHEN** a bolt reaches a Warrior with a shield facing the caster and the block roll succeeds
- **THEN** the bolt does no damage and both sides see it blocked

### Requirement: Damage over time ticks
A damage-over-time spell SHALL hit once for its `Damage` after the resist, every `Frequency` milliseconds, for `Duration` seconds, with no variance. A second cast of the same spell on the same target SHALL restart its duration, not stack. Status: reference.

#### Scenario: Five ticks
- **WHEN** a spell with Damage 10, Frequency 4000 and Duration 20 lands on a target with no resist
- **THEN** the target takes 10 five times, 4 seconds apart

#### Scenario: Recast
- **WHEN** the same spell lands again on that target after two ticks
- **THEN** one effect remains, with 20 seconds to run

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
A life-drain spell SHALL do damage as direct damage does, and SHALL heal its caster by that damage × `LifeDrainReturn` / 100, up to the caster's maximum health. Status: provisional, until the unit of `LifeDrainReturn` is confirmed.

#### Scenario: Drain
- **WHEN** a drain with Damage 40, return 50, rolls a variance of 1.0 on a target with no resist
- **THEN** the target loses 40 and the caster regains 20

### Requirement: Heals restore health
A heal spell SHALL restore `Value` health to its target, up to its maximum. Status: reference.

#### Scenario: A creature heals itself
- **WHEN** a creature at 100 of 300 health casts a 120-point heal on itself
- **THEN** it has 220

### Requirement: Effects show and end
Each lasting effect SHALL be sent to its target with its spell name, icon id and time left, and SHALL be removed when it ends, when its caster's spell is replaced, or when its target dies. Effects SHALL NOT survive a logout. Status: provisional for the logout rule.

#### Scenario: Effect icons
- **WHEN** a slow lands on a player
- **THEN** their client shows an effect with the spell name and a countdown, gone when the slow ends
