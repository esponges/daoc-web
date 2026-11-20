## Purpose

Defines stealth for the Shadowblade (and the Hunter's Camouflage): who can see a hidden player, what breaks stealth, how fast a hidden player moves, and the styles and abilities that depend on it.

## ADDED Requirements

### Requirement: A player with Stealth can hide
A player with the Stealth spec SHALL be able to enter stealth when not in combat. Doing so in combat SHALL be refused. A hidden player SHALL see themselves drawn translucent. Status: reference.

#### Scenario: Hide out of combat
- **WHEN** a Shadowblade not in combat uses stealth
- **THEN** they are hidden and their own client draws them translucent

#### Scenario: Hide in combat
- **WHEN** a Shadowblade who was attacked within the combat window uses stealth
- **THEN** it is refused with "You can't hide while in combat!"

### Requirement: Others see a hidden player only within detection range
A hidden player SHALL NOT be sent to other players or noticed by creatures beyond a detection range. The range SHALL come from the viewer's level and the hider's Stealth spec under the rule in the rule book, and SHALL be larger for a viewer with Detect Hidden. Inside the range, the hider SHALL be visible to that viewer and noticeable by an aggressive creature as normal. Status: set by the findings.

#### Scenario: Sneaking past an aggressive creature
- **WHEN** a hidden Shadowblade passes outside a creature's detection range but inside its aggro range
- **THEN** the creature does not attack

#### Scenario: Too close
- **WHEN** the Shadowblade comes inside the creature's detection range
- **THEN** the creature notices them and attacks as an aggressive creature would

#### Scenario: Creature detection
- **WHEN** a level 10 creature looks for a hider with Stealth 5
- **THEN** its detection radius is 225 units, 126 at least, and 125 more if it detects the hidden and the hider has no counter

#### Scenario: Too close to hide
- **WHEN** a player with Stealth 10 tries to hide with a level 40 enemy 500 units away
- **THEN** it is refused, because the limit is 1024 less 896 times 10 over 40, which is 800 units

#### Scenario: Another player
- **WHEN** a second player stands beyond their detection range of a hidden Shadowblade
- **THEN** the second client receives no position for the Shadowblade

### Requirement: Attacking, casting or being hurt ends stealth
Stealth SHALL end when the hider attacks, casts, uses an item or ability other than those allowed in stealth, or takes damage. Status: reference.

#### Scenario: Opening from stealth
- **WHEN** a hidden Shadowblade swings at a creature
- **THEN** stealth ends as the swing lands, and the swing may use a stealth-opener style

#### Scenario: Damage
- **WHEN** a hidden Shadowblade is hit by a damage-over-time tick
- **THEN** stealth ends

### Requirement: Hidden movement is slower
A hidden player SHALL move at a fraction of their normal speed, set by their Stealth spec under the rule, and the client's predicted movement SHALL agree with the server at that speed. Status: set by the findings.

#### Scenario: Stealth speed
- **WHEN** a hidden Shadowblade runs
- **THEN** they move at the stealth speed the rule gives, and the shared movement vector for that speed passes in Go and JS

### Requirement: The Shadowblade's other stealth abilities
Danger Sense, Detect Hidden, Distraction, Safe Fall and Climb Walls SHALL each be granted at the levels the data gives. Each SHALL do what the rule book records for PvE play, or be marked in the census as having no PvE effect, with the reason. Status: set by the findings.

#### Scenario: Detect Hidden
- **WHEN** a Shadowblade with Detect Hidden and one without stand at the same distance from a hidden player
- **THEN** only the one with Detect Hidden sees them, when that distance is inside the larger range only

### Requirement: Camouflage hides a Hunter from creatures' notice
A Hunter with Camouflage SHALL be able to use it to stay hidden under the same detection rule as stealth, with the differences the findings give. Status: set by the findings.

#### Scenario: Camouflaged Hunter
- **WHEN** a camouflaged Hunter passes outside an aggressive creature's detection range
- **THEN** the creature does not attack
