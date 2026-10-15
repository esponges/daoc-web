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
