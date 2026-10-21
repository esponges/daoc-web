## ADDED Requirements

### Requirement: A spellbook per character
A character of a casting class SHALL have a spellbook holding the spells of its spell lines up to its level, taken from the imported data. A spell SHALL appear in the book at the level the data gives it, and a level-up SHALL add the new spells and tell the player.

#### Scenario: A new Healer
- **WHEN** a level 1 Healer enters the world
- **THEN** the spellbook lists the spells of their lines whose level is 1 or less, and none above

#### Scenario: Levelling
- **WHEN** the Healer reaches the level of another spell in their lines
- **THEN** the spell is added to the book and the chat log says so

### Requirement: Casting
The client SHALL send a cast request naming a spell and, when the spell needs one, a target. The server SHALL accept it only when the caster is alive, knows the spell, has the power, is in range and in line of sight where the world has collision, and is not already casting. A cast SHALL take the spell's cast time, after which the effect applies and power is spent.

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
The server SHALL apply these spell types from the data: direct damage, bolt, heal, damage over time, damage add, strength, constitution, dexterity and quickness buffs, base armour buff, speed decrease, mesmerize and stun. Spell types outside the list SHALL stay in the book marked as not yet working, and casting one SHALL say so and spend nothing.

#### Scenario: A heal on a group member
- **WHEN** a Healer casts a heal on a group member in range
- **THEN** the member's health rises by the spell's amount, up to their maximum, and both see it

#### Scenario: A buff expires
- **WHEN** a buff's duration ends
- **THEN** the stat returns and the effect leaves the effects list

#### Scenario: An unsupported type
- **WHEN** a player casts a summon-minion spell
- **THEN** nothing is spent and the chat log says that spell does not work yet

### Requirement: The SPELLS window and quickbar
The stock SPELLS window SHALL list the spellbook by line, and a spell SHALL be placeable on the quickbar and cast from it. A cast bar SHALL show progress while a spell is cast. Without the converted interface the built-in panels SHALL list and cast spells.

#### Scenario: Cast from the bar
- **WHEN** the player drags a spell onto the bar and presses its key
- **THEN** the spell is cast at the current target

#### Scenario: No converted UI
- **WHEN** `web/data/ui/` is absent
- **THEN** the built-in panel still lists the spells and casts them
