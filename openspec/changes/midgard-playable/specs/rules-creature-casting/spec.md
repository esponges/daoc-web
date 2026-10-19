# Spec Delta

## Purpose

Defines when a creature whose template lists spells casts one, how long casting takes, what interrupts it, and what players see while it happens.

## ADDED Requirements

### Requirement: Creatures cast the spells their template lists
A creature SHALL know the spells its template lists that are of a type the server implements. Other listed spells SHALL be ignored and counted by the census. A creature with no implemented spells SHALL fight as before.

#### Scenario: A caster template
- **WHEN** a creature's template lists a direct-damage spell and a buff of an unimplemented type
- **THEN** it can cast the direct-damage spell and never casts the buff

### Requirement: A creature in combat casts when a spell is ready
At each of its combat decisions, a creature with a ready hostile spell whose range reaches its target SHALL cast it instead of swinging. A spell is ready when its recast delay has passed since its last cast. If none is ready, it SHALL swing or close in as before. Status: provisional.

#### Scenario: Cast, then melee
- **WHEN** a creature with one direct-damage spell (recast 30 s) aggroes a player 1,000 units away and the spell's range is 1,500
- **THEN** it casts the spell first, then closes and swings until 30 seconds have passed

#### Scenario: Out of range
- **WHEN** the target is beyond every ready spell's range
- **THEN** the creature closes in and does not cast until a spell reaches

### Requirement: Creatures heal when hurt
A creature with a heal SHALL cast it on itself, or on a hurt creature of its faction within the heal's range, when that creature is below 75% health and the heal is ready. A heal SHALL take priority over a hostile spell. Status: provisional.

#### Scenario: Healing a packmate
- **WHEN** a healer creature's packmate drops below 75% health within range
- **THEN** the healer casts its heal on the packmate

### Requirement: Casting takes the spell's cast time
A spell with a cast time SHALL take effect when its `CastTime` has passed. The caster SHALL stand still and face its target while casting. If the target is out of range or dead when the time ends, the spell SHALL fail and be ready again after 2 seconds. A cast time of 0 SHALL be instant. Status: reference, and provisional for the 2 seconds.

#### Scenario: A 3-second cast
- **WHEN** a creature begins a spell with CastTime 3000
- **THEN** it stands still for 3 seconds and the spell takes effect then

#### Scenario: The target ran away
- **WHEN** the target leaves the spell's range before the cast ends
- **THEN** the spell fails and the creature closes in

### Requirement: Being attacked interrupts casting
A melee attack on a casting creature SHALL interrupt its spell, whether the attack hits or not, unless the spell is marked uninterruptible. An interrupted spell SHALL have no effect and SHALL be ready again after 2 seconds. Status: reference, and provisional for the 2 seconds.

#### Scenario: Interrupt
- **WHEN** a Warrior swings at a creature 1 second into its 3-second cast
- **THEN** the cast stops, the creature's spell has no effect, and both see that it was interrupted

#### Scenario: Uninterruptible
- **WHEN** the spell is marked uninterruptible
- **THEN** the attack does not stop the cast

### Requirement: Players see a creature cast
When a creature begins a cast, nearby clients SHALL be told who is casting, the spell's name, its target and its cast time. The client SHALL play the creature's cast animation and SHALL print "<creature> begins casting a spell!". As on a 1.65 server, no cast bar SHALL be shown for another's cast.

#### Scenario: Watching a cast
- **WHEN** a creature begins casting at a player
- **THEN** that player and a nearby player see the line in chat and the creature's cast animation, and neither sees a cast bar
