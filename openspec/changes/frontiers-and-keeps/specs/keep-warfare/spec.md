## ADDED Requirements

### Requirement: Claiming
A guild with a group of its members inside a keep its realm owns and nobody has claimed SHALL be able to claim it. The guild's banners SHALL show. The guild SHALL be able to release it.

#### Scenario: Claim
- **WHEN** a group of "Hammers of Odin" members claims a captured keep
- **THEN** the keep shows their banner and the Keep window names the guild

### Requirement: Upgrades
A claiming guild SHALL be able to raise a keep's level for the rule book's cost and time. Guards and doors SHALL be stronger by level.

#### Scenario: Upgrade
- **WHEN** a guild upgrades its keep from level 1 to 2 and the time passes
- **THEN** the keep is level 2 and its doors have the level 2 health

### Requirement: Siege engines
A player SHALL be able to place a siege engine they hold at a valid ground target, then aim, load and fire it. Its ammunition SHALL damage doors and walls, or players in its area for anti-personnel ammunition. Engines SHALL be repairable and destructible.

#### Scenario: Ram on a door
- **WHEN** a ram is placed at an enemy door and operated
- **THEN** each strike damages the door by the rule book's amount

#### Scenario: Not near a door
- **WHEN** a player places a ram away from any door
- **THEN** it is refused

### Requirement: Repair and hookpoints
Owners with woodworking skill SHALL be able to repair doors and walls with materials. Defenders SHALL be able to buy guards and boiling oil at hookpoints.

#### Scenario: Repairing a door
- **WHEN** an owner with woodworking repairs a damaged door with wood
- **THEN** the door's health rises by the rule book's amount

### Requirement: Keep teleports
Owners SHALL be able to teleport between their realm's connected keeps through the keep's portal stone.

#### Scenario: Keep to keep
- **WHEN** a Midgard player uses the stone in a Midgard-held keep
- **THEN** they can choose another connected Midgard-held keep and arrive there

### Requirement: Climbing walls
A character with Climb Walls SHALL be able to climb keep walls at climb points.

#### Scenario: Climbing
- **WHEN** a Shadowblade uses Climb Walls at a keep wall
- **THEN** they arrive on the wall top
