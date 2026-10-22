# Spec Delta

## Purpose

Gives Midgard its 1.65 travel: stable masters sell tickets, and handing one back puts the player on a horse that follows the imported route to another town.

## ADDED Requirements

### Requirement: Stable masters sell tickets
A stable master SHALL sell the tickets its imported stock list holds, through the merchant window, for the listed price.

#### Scenario: Buying a ticket
- **WHEN** a player right-clicks the stable master in Mularn and buys the ticket to Audliten
- **THEN** the ticket is in their pack and its price is taken

### Requirement: A ticket starts its route
A player who gives a stable master a ticket whose route starts at that stable SHALL be mounted and ridden along the route. The ticket SHALL be used up. A ticket for a route that starts elsewhere SHALL be refused and kept. The route SHALL be the imported path the ticket names.

#### Scenario: Riding out
- **WHEN** a player hands the Mularn stable master the ticket to Audliten
- **THEN** the ticket is gone and the player is on a horse following the path from Mularn to Audliten

#### Scenario: The wrong stable
- **WHEN** the player hands that ticket to the stable master in Audliten
- **THEN** it is refused and stays in their pack

### Requirement: The ride follows the path
The rider SHALL move through the path's points in order, at each point's speed or the path's default horse speed, crossing zone borders as walking would. They SHALL face the direction of travel. At the last point they SHALL be dismounted where they stand. The movement SHALL be decided by the server. Status: reference for the speed default.

#### Scenario: Arrival
- **WHEN** the horse reaches the route's last point
- **THEN** the player is dismounted there and can move again

#### Scenario: Across a border
- **WHEN** the route crosses from Vale of Mularn into another zone
- **THEN** the rider enters that zone without stopping

### Requirement: Riders do not fight
A rider's movement keys SHALL do nothing, and they SHALL NOT attack, use styles or abilities, or be attacked by creatures. Creatures SHALL NOT aggro on a rider. Chat SHALL work. Status: reference.

#### Scenario: Past an aggressive creature
- **WHEN** the route passes within aggro range of an aggressive creature
- **THEN** the creature does not attack the rider

### Requirement: Others see the horse
Every client near a rider SHALL see them seated on a horse moving along the route. Where the horse model is not converted, the rider SHALL be drawn moving with their run animation.

#### Scenario: A friend watches
- **WHEN** a second player stands beside the route as a rider passes
- **THEN** they see the rider on a horse moving along it
