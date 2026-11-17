## MODIFIED Requirements

### Requirement: Spell types of the first slice
The server SHALL apply these spell types from the data: direct damage, bolt, heal, damage over time, damage add, strength, constitution, dexterity and quickness buffs, base armour buff, speed decrease, mesmerize, stun and life drain. A stat buff SHALL add the spell's `Value` whatever the caster's stat or spec, and the stats a player gets from buffs SHALL be at most the caster-level cap of 1.25 per level. Buffs of one kind do not stack: the strongest counts. Spell types outside the list SHALL stay in the book marked as not yet working, and casting one SHALL say so and spend nothing. Status: reference (DOLSharp and OpenDAoC-Core agree on the flat value and the cap).

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

## ADDED Requirements

### Requirement: The armour-factor buff cap
The armour factor a player gets from buff spells SHALL be at most twice their level. Status: provisional. Neither community server states a flat cap for the base armour buff (DOLSharp caps specialisation-line armour buffs at 1.875 per level, and OpenDAoC-Core applies base armour buffs piece by piece), and no 1.65 account was found.

#### Scenario: A big armour buff
- **WHEN** a level 10 player holds armour buffs worth 50
- **THEN** their armour factor from buffs is 20
