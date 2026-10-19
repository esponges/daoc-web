# Spec Delta

## MODIFIED Requirements

### Requirement: Experience and levels
Killing NPCs SHALL grant experience as `rules-experience` defines. Gaining a level SHALL raise the character's stats, hit points and specialisation points as `rules-character-stats` and `rules-specialization` define. Styles SHALL unlock by spec level, not by character level.

#### Scenario: Level up
- **WHEN** a character's experience crosses the threshold for the next level
- **THEN** the server raises the level, announces it with the stats, hit points and spec points gained, and resends the character sheet

#### Scenario: Grey kills
- **WHEN** a player kills an NPC that cons grey to them
- **THEN** they gain no experience

### Requirement: What a Warrior may carry and wear
The server SHALL let a character equip only what their class's abilities allow, at their current level. For a Warrior that means:
- swords, axes and hammers, one- or two-handed, from the weapon families they have;
- a shield no larger than their Shield ability allows, and only with a one-handed weapon or none;
- armour no heavier than their armour ability: studded until level 10, chain from 10.

Items SHALL be from Midgard or no realm, and SHALL NOT have a level requirement above the character's level. Anything else SHALL be refused with a reason.

#### Scenario: A Hibernian blade
- **WHEN** a player tries to equip a Hibernian weapon
- **THEN** it is refused and their weapon is unchanged

#### Scenario: A pelt on the head
- **WHEN** a player tries to wear an item in a slot it does not belong to
- **THEN** it is refused and what was worn there stays

#### Scenario: A great axe
- **WHEN** a Warrior with a shield equips a two-handed axe
- **THEN** the axe is equipped, the shield goes to the pack (or the axe is refused if the pack is full), and the style bar is rebuilt for two-handed axe styles

#### Scenario: Chain too early
- **WHEN** a level 8 Warrior tries to wear a chain hauberk
- **THEN** it is refused because they cannot wear chain until level 10

#### Scenario: A high-level weapon
- **WHEN** a level 5 Warrior equips a level 20 sword that has no level requirement
- **THEN** it is equipped, and its damage per second counts only up to the level 5 cap

### Requirement: Equipment sets the numbers, not the picture
A character's blows, swing time, armour, stats and styles SHALL come from the items they wear and their bonuses, within the caps of the rules. Their drawn appearance SHALL stay their outfit's.

#### Scenario: Swapping hammers
- **WHEN** a player equips a heavier hammer
- **THEN** their swing damage and timer change and the style bar is rebuilt for the weapon type

#### Scenario: A stat bonus
- **WHEN** a player wears gloves with +3 dexterity
- **THEN** their dexterity, and so their parry and block chances, rise by that amount, within the item cap
