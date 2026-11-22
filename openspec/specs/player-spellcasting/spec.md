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

### Requirement: The SPELLS window and quickbar
The SPELLS window SHALL be the stock `stats_spells` window, listing the spellbook in its tree, with one heading per spell line and each spell's icon by kind, name and level. A spell SHALL be castable from the window and placeable on the quickbar by dragging, and cast from there. The cast bar SHALL be the stock `Timer` window, showing the spell's name and its progress while the spell is cast. Without the converted interface the built-in panels SHALL list and cast spells.

#### Scenario: Cast from the bar
- **WHEN** the player drags a spell onto the bar and presses its key
- **THEN** the spell is cast at the current target

#### Scenario: The spell tree
- **WHEN** a Bonedancer opens SPELLS
- **THEN** the stock window shows headings for Bone Army, Darkness and Suppression, and expanding one lists its spells with their icons

#### Scenario: The cast bar
- **WHEN** the player begins a three-second cast
- **THEN** the stock Timer window shows the spell's name and fills over three seconds, and hides when the spell lands or is interrupted

#### Scenario: No converted UI
- **WHEN** `web/data/ui/` is absent
- **THEN** the built-in panel still lists the spells and casts them

### Requirement: The armour-factor buff cap
The armour factor a player gets from buff spells SHALL be at most twice their level. Status: provisional. Neither community server states a flat cap for the base armour buff (DOLSharp caps specialisation-line armour buffs at 1.875 per level, and OpenDAoC-Core applies base armour buffs piece by piece), and no 1.65 account was found.

#### Scenario: A big armour buff
- **WHEN** a level 10 player holds armour buffs worth 50
- **THEN** their armour factor from buffs is 20

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

### Requirement: Player casts respect line of sight
A player's spell aimed at another combatant SHALL be checked for line of sight when it starts and when it lands, under `rules-line-of-sight`. A spell that fails the check SHALL spend no power and start no recast timer.

#### Scenario: Blocked at the start
- **WHEN** a Shaman starts a damage-over-time spell at a creature they cannot see
- **THEN** the cast does not start, "You can't see your target!" is shown, and the spell stays ready

### Requirement: Resurrection spells raise a dead player
A resurrection spell SHALL target a dead player who has not released, within its range. The dead player SHALL be offered the resurrection and SHALL accept or decline it. When they accept, they SHALL rise where they lie, with the health and power percentages the spell gives. A spell aimed at a living or released player SHALL be refused.

#### Scenario: Raising a group member
- **WHEN** a Healer casts a resurrection with 10% health on a dead group member, who accepts
- **THEN** the member stands up where they died with 10% of their health and power, and is no longer dead

#### Scenario: Declined
- **WHEN** the dead player declines
- **THEN** they stay dead, may still release, and the caster's power is not refunded

#### Scenario: Already released
- **WHEN** the target has already released
- **THEN** the spell is refused with "Your target is not dead" or the game's message for it

### Requirement: Concentration buffs are cast like other buffs
A spell with a concentration cost SHALL be cast through the same checks as any buff, plus the concentration check in `rules-concentration`. The spell bar and SPELLS window SHALL show a held concentration buff as on.

#### Scenario: Buff bot behaviour
- **WHEN** a Shaman casts their strength/constitution concentration buff on each of three group members
- **THEN** each holds the buff with no timer, and the Shaman's concentration in use is three times its cost

### Requirement: Summon and charm spells make pets
Summon spells SHALL create a pet, and charm spells SHALL take a creature as a pet, under `pets`. A minion summon with no commander SHALL be refused.

#### Scenario: A Hunter's pet
- **WHEN** a Hunter casts their Beastcraft summon
- **THEN** their pet appears and the pet window opens

### Requirement: Spells aimed at the caster's pet can be cast
A spell whose target is the pet SHALL fall on the caster's own pet (a Bonedancer's commander; never a minion), and be refused with "You must have a pet to cast that." when there is none. Every lasting effect a creature can hold (strength and constitution, dexterity and quickness, combat speed, damage add, damage shield, ablative armour, mesmerize duration, an offensive proc) SHALL work on a pet in its own fights, and a heal SHALL raise it.

#### Scenario: Pet spell without a pet
- **WHEN** a Spiritmaster with no pet casts a buff aimed at their pet
- **THEN** the cast is refused and no power is spent

#### Scenario: A pet proc
- **WHEN** a pet holding an offensive proc strikes a creature and the proc's chance comes up
- **THEN** the proc's spell is cast on the creature by the pet

### Requirement: Debuffs, cures and the remaining Midgard buffs can be cast
Every spell type that the twelve Midgard classes' lines use SHALL be castable by a player, under the effects in `rules-spell-effects`: the remaining buffs, the debuffs and control debuffs, the cures and spread heal, life transfer, and damage with a debuff or slow. Any type left over SHALL be counted in the census with a reason.

#### Scenario: A Bonedancer's strength debuff
- **WHEN** a Bonedancer casts a strength debuff on a creature that does not resist
- **THEN** the creature's strength falls by the spell's value for its duration, and its melee damage falls with it

#### Scenario: Cure Poison
- **WHEN** a Healer casts Cure Poison on a group member carrying a poison
- **THEN** the poison effect ends on them

### Requirement: One press asks for one cast
A double-click on a spell, or a key held down on a quickbar slot, SHALL send one cast request. A second request SHALL be sent only for a separate press made after the first cast has ended, failed or been interrupted.

#### Scenario: Double-click
- **WHEN** the player double-clicks Minor Dusk Dart in the spells window
- **THEN** one cast begins and no "You are already casting a spell." line appears

#### Scenario: Held key
- **WHEN** the player holds a quickbar key down through a cast
- **THEN** one cast request is sent
