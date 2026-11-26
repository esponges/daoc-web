# server-npcs Specification

## Purpose
TBD - created by archiving change multiplayer-server. Update Purpose after archive.

## Requirements

### Requirement: Spawns come from imported data
The server SHALL populate each zone from the imported spawn table, using each spawn's template for name, model, level, size and equipment.

#### Scenario: Zone 100 population
- **WHEN** the server loads zone 100 with imported data
- **THEN** its NPCs are the imported spawns, not the hand-written `web/spawns/zone100.json`

#### Scenario: Offline viewer
- **WHEN** the viewer runs without a server
- **THEN** it uses `web/spawns/zone100.json` as before

### Requirement: NPCs wander, aggro and leash
The server SHALL move NPCs along their imported paths or within their roam radius at walking pace on the ground, attack players that enter their aggro range when aggressive, chase to a leash distance and then return home and heal.

#### Scenario: Aggressive wolf
- **WHEN** a player enters an aggressive wolf's aggro range
- **THEN** the wolf closes at its run speed and attacks

#### Scenario: Packmates
- **WHEN** a player attacks a creature, or is noticed by one
- **THEN** others of the same name join the fight only if they are aggressive, the creature is within their own aggro range, and the player is not grey to them; a passive creature fights only when it is hit itself. Status: provisional

#### Scenario: Leash
- **WHEN** a player draws an NPC beyond its leash distance
- **THEN** it gives up, walks home and heals

#### Scenario: Feet stay dry
- **WHEN** an NPC wanders
- **THEN** it does not enter water deeper than its walking limit

### Requirement: Death and respawn
A killed NPC SHALL lie for a time, vanish, and respawn at home after its respawn interval.

#### Scenario: Respawn
- **WHEN** an NPC is killed
- **THEN** after its corpse time and respawn interval a fresh NPC stands at its home with full health

### Requirement: Every client sees the same NPC
All clients in a zone SHALL see an NPC at the same place and in the same state, within snapshot latency.

#### Scenario: Two observers
- **WHEN** two players watch the same wolf
- **THEN** it is in the same place on both screens

### Requirement: Friendly creatures never fight
A creature whose spawn belongs to a realm (guards, villagers, merchants) SHALL wander within its roam, SHALL NOT attack or aggro, and SHALL NOT be attackable.

#### Scenario: A villager in Mularn
- **WHEN** a player targets a villager and switches attack on
- **THEN** the server refuses and tells them they cannot attack it

### Requirement: Guards defend their realm
A guard SHALL attack an aggressive creature of another realm that comes within the guard's aggro range, or the imported default of 750 units when the guard has none. It SHALL also attack such a creature that is fighting a player of the guard's realm within that range. It SHALL chase it up to its leash, as other NPCs do, and SHALL then return to its post. Status: reference.

#### Scenario: A creature follows a player into town
- **WHEN** an aggressive wolf chases a Midgard player past a Mularn guard within its range
- **THEN** the guard attacks the wolf, and the wolf turns on whoever it hates most

#### Scenario: Friendly creatures are left alone
- **WHEN** a non-aggressive creature wanders near a guard
- **THEN** the guard does not attack it

### Requirement: Guard kills give no experience to players
A creature killed with damage from a guard SHALL count the guard's damage as an unshared damager. Players SHALL get only their own share of the damage. Status: reference.

#### Scenario: A guard does half
- **WHEN** a guard deals 50% of a level 10 creature's damage and a level 10 Warrior deals 50%
- **THEN** the Warrior gains 2,560 and the other half is given to no one

### Requirement: Dropped placements are accounted for
A placement the world does not contain SHALL be explained. The README SHALL state how many Midgard placements lack a converted model and why, so a skipped creature is a recorded limitation and not a silent one.

#### Scenario: Documented gap
- **WHEN** a reader looks for creatures such as wisps, skugga and mud frogs in Midgard
- **THEN** the README's known limitations say their models are effect or composite models the character converter cannot yet read, with the count of placements affected
