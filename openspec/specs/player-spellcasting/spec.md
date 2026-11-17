# player-spellcasting Specification

## Purpose
TBD - created by archiving change next-midgard-classes-and-spells. Update Purpose after archive.

## Requirements

### Requirement: A spellbook per character
A character of a casting class SHALL have a spellbook holding the spells of the spell lines of the specs it trains, taken from the imported data. A spell SHALL appear in the book when the spec reaches the level the data gives it, and training a spec SHALL add the new spells and tell the player.

#### Scenario: A new Healer
- **WHEN** a level 1 Healer enters the world
- **THEN** the spellbook lists the spells of their specs' lines whose level is 1 or less, and none above

#### Scenario: Levelling
- **WHEN** the Healer trains a spec to the level of another spell in its line
- **THEN** the spell is added to the book and the chat log says so

### Requirement: Casting
The client SHALL send a cast request naming a spell and, when the spell needs one, a target. The server SHALL accept it only when the caster is alive, knows the spell, has the power, is in range, and is not already casting. Line of sight is not checked yet. A cast SHALL take the spell's cast time, after which the effect applies and power is spent.

#### Scenario: Out of power
- **WHEN** a caster asks for a spell that costs more power than they have
- **THEN** the server refuses and the chat log says "You don't have enough power"

#### Scenario: Out of range
- **WHEN** the target is beyond the spell's range
- **THEN** the cast does not start and the chat log says the target is too far away

#### Scenario: A bolt
- **WHEN** a Runemaster casts a bolt at a creature in range
- **THEN** the cast bar fills over the cast time, the creature takes damage and turns on the caster

### Requirement: Interrupts and recast
A cast SHALL be interrupted when the caster takes damage from a melee blow or spell, moves, or is stunned or mesmerized; an interrupted cast spends no power and is followed by the interrupt delay. A spell with a recast time SHALL not be castable again until it ends.

#### Scenario: Hit while casting
- **WHEN** a creature hits a caster mid-cast
- **THEN** the cast stops, the cast bar clears and the chat log says "You are interrupted"

#### Scenario: Recast
- **WHEN** a spell with a recast time was just cast
- **THEN** a second request is refused until the time passes

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

### Requirement: The SPELLS window and quickbar
The stock SPELLS window SHALL list the spellbook by line, and a spell SHALL be placeable on the quickbar and cast from it. A cast bar SHALL show progress while a spell is cast. Without the converted interface the built-in panels SHALL list and cast spells.

#### Scenario: Cast from the bar
- **WHEN** the player drags a spell onto the bar and presses its key
- **THEN** the spell is cast at the current target

#### Scenario: No converted UI
- **WHEN** `web/data/ui/` is absent
- **THEN** the built-in panel still lists the spells and casts them

### Requirement: The armour-factor buff cap
The armour factor a player gets from buff spells SHALL be at most twice their level. Status: provisional. Neither community server states a flat cap for the base armour buff (DOLSharp caps specialisation-line armour buffs at 1.875 per level, and OpenDAoC-Core applies base armour buffs piece by piece), and no 1.65 account was found.

#### Scenario: A big armour buff
- **WHEN** a level 10 player holds armour buffs worth 50
- **THEN** their armour factor from buffs is 20
