## MODIFIED Requirements

### Requirement: Spell types of the first slice
The server SHALL apply these spell types from the data: direct damage, bolt, heal, damage over time, damage add, strength, constitution, dexterity and quickness buffs, base armour buff, speed decrease, mesmerize, stun, life drain, speed enhancement, health regeneration buff, and the body, spirit, energy, heat, cold and matter resist buffs. A stat buff SHALL add the spell's `Value` whatever the caster's stat or spec, and the stats a player gets from buffs SHALL be at most the caster-level cap of 1.25 per level. Buffs of one kind do not stack: the strongest counts. Spell types outside the list SHALL stay in the book marked as not yet working, and casting one SHALL say so and spend nothing. Status: reference (DOLSharp and OpenDAoC-Core agree on the flat value and the cap).

#### Scenario: A heal on a group member
- **WHEN** a Healer casts a heal on a group member in range
- **THEN** the member's health rises by the spell's amount, up to their maximum, and both see it

#### Scenario: A buff expires
- **WHEN** a buff's duration ends
- **THEN** the stat returns and the effect leaves the effects list

#### Scenario: An unsupported type
- **WHEN** a player casts a summon-minion spell
- **THEN** nothing is spent and the chat log says that spell does not work yet

#### Scenario: A buff is capped by level
- **WHEN** a level 10 player holds buffs worth 20 strength
- **THEN** their strength rises by 12

#### Scenario: A buff does not scale with the caster
- **WHEN** a caster with 50 or with 200 in their casting stat casts a 20-point strength buff
- **THEN** the target's strength rises by the same amount in both cases

#### Scenario: Resist buffs of one kind
- **WHEN** a player holds a 10-point and a 20-point heat resist buff
- **THEN** their heat resist rises by 20

## ADDED Requirements

### Requirement: Area spells hit every enemy inside the radius
A player's hostile spell with a radius SHALL land on every living, hostile creature within that radius of its centre. The centre SHALL be the caster when the spell's range is 0, and the target otherwise. Each creature SHALL roll its own hit and resist, and SHALL be put in combat with the caster. Players SHALL never be hit. A spell with range 0 SHALL need no target.

#### Scenario: Thunder Shout
- **WHEN** a Thane casts Thunder Shout (range 0, radius 300) with three creatures 200 units away and one 400 units away
- **THEN** the three near creatures each take a damage roll and attack the Thane, and the far one is untouched

#### Scenario: Mjollnir around the target
- **WHEN** a Thane casts Call Mjollnir at a creature that stands beside another
- **THEN** both creatures are hit, and a group member standing beside them is not

#### Scenario: No target needed
- **WHEN** a Thane with no target casts Thunder Shout
- **THEN** the spell is cast, and hits whatever is in range or nothing

### Requirement: Songs and chants pulse while on
A spell with a pulse SHALL be a toggle. Casting it SHALL turn it on and apply it at once to the caster's group in range, and again every pulse interval after (the spell's frequency). Each application SHALL last a pulse interval plus one second, so the effect lapses on a member who leaves range. Casting it again SHALL turn it off. Only one song may be on at a time: starting another SHALL end the first. Death SHALL turn it off; a held caster SHALL skip pulses and keep the song. Status: reference.

#### Scenario: A Song of Travel
- **WHEN** a Skald in a group of three turns on Simple Song of Travel
- **THEN** all three run faster within a moment, and keep doing so while the song is on and they stay within 2000 units

#### Scenario: Leaving range
- **WHEN** a group member walks 3000 units away from a Skald whose song is on
- **THEN** within six seconds the effect leaves that member, and returns when they come back in range

#### Scenario: Toggling off
- **WHEN** the Skald casts the same song again
- **THEN** the song stops pulsing, its effect ends on everyone within six seconds, and the spell bar no longer shows it as on

#### Scenario: A second song
- **WHEN** a Skald with Chant of the Brawl on casts Song of Rest
- **THEN** the chant is off and the song is on

#### Scenario: The caster dies
- **WHEN** a Skald dies with a chant on
- **THEN** the chant is off

#### Scenario: A held caster
- **WHEN** a Skald is mesmerized for 12 seconds with a song on
- **THEN** no pulse reaches the group while they are held, and the song is still on when they are free
