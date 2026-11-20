# rules-line-of-sight Specification

## Purpose
Decides when one combatant can see another for spells and ranged attacks, using the world the server already loads (terrain, building and dungeon hulls, tree trunks), and what a caster or archer does when it cannot.

## Requirements

### Requirement: Line of sight is a clear segment between eyes
The server SHALL treat a target as visible when the segment between the two combatants' eye points crosses no terrain, no blocking hull triangle and no tree trunk. The eye point is the body's position raised by a fixed eye height. Status: provisional.

#### Scenario: Open ground
- **WHEN** a player and a creature stand 1000 units apart on flat ground with nothing between them
- **THEN** each can see the other

#### Scenario: A hill between them
- **WHEN** a ridge of the heightmap rises above the segment between their eye points
- **THEN** neither can see the other

#### Scenario: A dungeon wall between them
- **WHEN** a player and a creature in a dungeon stand on either side of a wall of the interior's geometry
- **THEN** neither can see the other

#### Scenario: A tree trunk
- **WHEN** the segment passes through the trunk radius of a placed tree
- **THEN** the target is not visible

### Requirement: A player's hostile spell needs line of sight at start and at landing
A player SHALL NOT begin a spell aimed at another combatant they cannot see, and a spell whose target is no longer visible when the cast completes SHALL fail without its power being spent. Self, group and range-0 area spells SHALL need no line of sight to their centre, but each creature an area spell hits SHALL be visible from the centre. Status: reference.

#### Scenario: Casting at a creature behind a wall
- **WHEN** a Runemaster starts a bolt at a creature behind a dungeon wall
- **THEN** the cast is refused with the message "You can't see your target!" and no power is spent

#### Scenario: The target steps behind cover during the cast
- **WHEN** the target walks behind a building before a 3-second cast completes
- **THEN** the spell fails with the same message and the caster keeps their power

#### Scenario: Area spell around the caster
- **WHEN** a Thane uses Thunder Shout with one creature in the open and one behind a wall inside the radius
- **THEN** only the creature in the open is hit

### Requirement: A creature without line of sight closes in instead of casting
A creature SHALL NOT cast a hostile spell at a combatant it cannot see. It SHALL move toward its target as for melee until it can see it, then cast as before. Status: reference.

#### Scenario: Caster creature behind a corner
- **WHEN** a casting creature's target stands around a dungeon corner within the spell's range
- **THEN** the creature does not cast, and moves toward the target until it sees it

### Requirement: Line of sight is cheap enough for every cast
A line-of-sight check SHALL use the zone's existing binned collision data and SHALL NOT load any geometry beyond what movement collision loads. A zone without collision data SHALL treat every target as visible. Status: provisional.

#### Scenario: A zone without collision data
- **WHEN** the server runs a zone whose collision data is absent
- **THEN** every line-of-sight check passes and casting works as before this change
