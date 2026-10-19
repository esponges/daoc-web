# Spec Delta

## Purpose

Gives the world a game clock that every client shares, with day and night, and makes the imported night-only spawns keep to night.

## ADDED Requirements

### Requirement: The server keeps a game clock
The server SHALL keep a game time of day in which one game day lasts 60 real minutes. The clock SHALL keep running while players are offline, so that game time follows real time. Every client SHALL receive the game time when entering the world and at least once a minute after. Status: reference.

#### Scenario: Two clients agree
- **WHEN** two players are in the world at once
- **THEN** their clocks show the same game time to within one game minute

#### Scenario: A quarter of an hour later
- **WHEN** 15 real minutes pass
- **THEN** the game clock has moved 6 game hours

### Requirement: Night is 18:00 to 06:00
Game times from 18:00 up to 06:00 SHALL be night, and the rest day. Status: reference.

#### Scenario: Dusk
- **WHEN** the game clock passes 18:00
- **THEN** it is night

### Requirement: Night-only spawns keep to night
Spawns the data marks as night-only SHALL be present only at night. At dusk they SHALL appear at their spawn points. At dawn they SHALL leave, unless they are in combat, in which case they SHALL leave when the fight ends. A night-only spawn killed at night SHALL respawn as usual only while it is night.

#### Scenario: Dusk in a zone
- **WHEN** night falls
- **THEN** that zone's night-only creatures appear and every client nearby sees them

#### Scenario: Dawn mid-fight
- **WHEN** dawn comes while a night-only creature is fighting a player
- **THEN** it stays until the fight ends, then leaves

### Requirement: Day and night are counted at load
The load report SHALL count the night-only spawns per zone.

#### Scenario: Load report
- **WHEN** the server starts
- **THEN** the log names how many night-only spawns each zone has

### Requirement: The client shows day and night
The client SHALL show the game time on the stock clock and SHALL light the sky by it: the converted clear day by day, and a darker night sky and light at night. Where night sky art is not converted, it SHALL darken the day sky.

#### Scenario: Night falls on the client
- **WHEN** the game clock passes 18:00
- **THEN** the clock shows the game time and the scene darkens over the next game hour
