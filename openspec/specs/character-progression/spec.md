# character-progression Specification

## Purpose
TBD - created by archiving change multiplayer-server. Update Purpose after archive.

## Requirements

### Requirement: Inventory and equipment persist
A character SHALL have an inventory and equipped items that the server stores, validates and restores.

#### Scenario: Equip an item
- **WHEN** a player equips a hammer from their inventory
- **THEN** the server checks the class may use it, moves it to the held slot, and every nearby client draws it in the hand

#### Scenario: Item the class cannot use
- **WHEN** a player tries to equip an item their class cannot use
- **THEN** the server refuses it

### Requirement: Killing drops loot
A killed NPC SHALL drop items and money according to its imported loot template, and only the killer, or their group, SHALL be able to take them.

#### Scenario: Loot a wolf
- **WHEN** a player kills a wolf with a loot template
- **THEN** the drops roll once on the server and the player can take them from the corpse

#### Scenario: Someone else's kill
- **WHEN** another player tries to loot the corpse in the exclusive window
- **THEN** the server refuses it

### Requirement: Experience and levels
Killing NPCs SHALL grant experience scaled by the NPC's level against the player's, and gaining a level SHALL raise the character's hit points and unlock the styles for that level.

#### Scenario: Level up
- **WHEN** a character's experience crosses the threshold for the next level
- **THEN** the server raises the level, announces it, and the character may use styles granted at that level

#### Scenario: Grey kills
- **WHEN** a player kills an NPC many levels below them
- **THEN** the experience is reduced or zero

### Requirement: Server-side validation of item changes
All item changes SHALL be made by the server and each SHALL be atomic with saving the character.

#### Scenario: Crash during pickup
- **WHEN** the server restarts after an item was picked up
- **THEN** the item is in the inventory once, not twice and not lost

### Requirement: What a Warrior may carry and wear
The server SHALL let a Warrior equip one-handed swords, hammers and axes, shields, and armour up to chain, from Midgard or no realm, no more than two levels above their own, each in the slot it belongs to, and SHALL refuse anything else with a reason.

#### Scenario: A Hibernian blade
- **WHEN** a player tries to equip a Hibernian weapon
- **THEN** it is refused and their weapon is unchanged

#### Scenario: A pelt on the head
- **WHEN** a player tries to wear an item in a slot it does not belong to
- **THEN** it is refused and what was worn there stays

### Requirement: Equipment sets the numbers, not the picture
A character's blows, swing time, armour and styles SHALL come from the items they wear, while their drawn appearance stays their outfit's.

#### Scenario: Swapping hammers
- **WHEN** a player equips a heavier hammer
- **THEN** their swing damage and timer change and the style bar is rebuilt for the weapon type
