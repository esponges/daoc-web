# Spec Delta

## Purpose

The rule book's entry for the movement numbers beyond walk and run: sprint speed and its endurance cost, swim speed, and the jump and gravity the shared movement step uses.

## ADDED Requirements

### Requirement: Sprint speed and cost
Sprinting while running forward SHALL multiply run speed by the sprint factor and spend endurance each second while the player moves. Standing still while sprint is on SHALL cost nothing. The sprint factor is 1.25 and the cost is 5 endurance a second spent moving. Status: provisional; no 1.65 page gives them (the Grab Bag question about sprint was never answered in what is published).

#### Scenario: Sprint speed
- **WHEN** a player sprints forward on flat ground with no other speed effect
- **THEN** they move at 191 × 1.25 = 238.75 units a second

#### Scenario: Standing sprint costs nothing
- **WHEN** a player turns sprint on and stands still for ten seconds
- **THEN** their endurance does not fall

#### Scenario: Out of endurance
- **WHEN** a sprinting player's endurance reaches zero
- **THEN** sprint turns off and they run at run speed

### Requirement: Swim speed
A swimming player SHALL move at the swim speed whatever the run flag, and speed buffs SHALL apply to it as to running. The swim speed is 120 units a second. Status: provisional.

#### Scenario: Swimming
- **WHEN** a player swims forward with no speed effect
- **THEN** they move at 120 units a second

### Requirement: Jump and gravity
A jump SHALL start with a fixed upward speed, and every body in the air SHALL fall under one fixed gravity. The jump speed is 200 units a second and gravity 500 units a second squared. Both are chosen so a jump on flat ground lasts as long as the client's jump and land clips. Status: project. No source gives the client's numbers, and they decide no outcome beyond falling damage.

#### Scenario: Airtime
- **WHEN** a player jumps on flat ground
- **THEN** they are in the air for the number of steps the spec's constants give, the same on server and client

### Requirement: Registered and tested
Each rule in this spec SHALL have a table test in `internal/server/rules` for each numbered scenario and a registry entry so that `cmd/parity` counts it, with its sources and status.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs
- **THEN** the movement rules are listed with their status
