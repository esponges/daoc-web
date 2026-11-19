## Purpose

Defines how a Hunter's bow attack works at 1.65: arrows as ammunition, drawing and releasing, range, line of sight, hit and damage, and the bow abilities Critical Shot, Rapid Fire and Penetrating Arrow.

## ADDED Requirements

### Requirement: A bow shot needs a bow, an arrow and a visible target in range
A ranged attack SHALL need:
- a bow in the ranged slot;
- at least one arrow of a kind the bow fires, in the quiver slots;
- a target that can be fought, within the bow's range and visible under line of sight.

Otherwise it SHALL be refused with the game's message. Each released shot SHALL use up one arrow. Status: reference.

#### Scenario: No arrows
- **WHEN** a Hunter with an empty quiver tries to shoot
- **THEN** the shot is refused with "You have no ammunition" and nothing is drawn

#### Scenario: Out of range
- **WHEN** the target is beyond the bow's range
- **THEN** the shot is refused with "That target is too far away!"

### Requirement: Drawing takes the bow's speed and can be interrupted
A shot SHALL draw for a time set by the bow's speed and the archer's quickness under the swing-speed rule. A draw SHALL be interrupted by the archer moving, and by taking damage under the rule the findings give. A finished draw SHALL release the arrow at once. The draw progress SHALL be shown on the cast bar. Status: reference.

#### Scenario: A full draw
- **WHEN** a Hunter draws a 4.5-speed composite bow standing still
- **THEN** the cast bar runs for the draw time the rule gives, and the arrow is released when it ends

#### Scenario: Moving
- **WHEN** the Hunter moves during the draw
- **THEN** the draw stops, and no arrow is used

### Requirement: An arrow hits and does damage like a swing from range
An arrow SHALL roll hit, evade and block (no parry) against the target as a swing does. On a hit, it SHALL do damage from the bow's DPS and speed, the archer's composite bow spec and dexterity, the arrow's kind and the target's armour, by the formula recorded in the rule book. Status: set by the findings.

#### Scenario: A worked example
- **WHEN** a level 20 Hunter with composite bow 20 hits a level 20 creature with a 12.0 DPS, 4.5-speed bow
- **THEN** the damage equals the rule book's worked example for those numbers

#### Scenario: Parry does not apply
- **WHEN** an arrow is shot at a creature with a parry chance
- **THEN** the shot is never parried

### Requirement: Critical Shot, Rapid Fire and Penetrating Arrow change a shot
The bow abilities SHALL each change one shot:
- **Critical Shot** draws longer and adds damage by its level;
- **Rapid Fire** shortens the draw and lowers damage;
- **Penetrating Arrow** lets a shot pass through a bladeturn.

Each SHALL be learned at the composite bow levels the data grants, by the rule in the rule book. Status: set by the findings.

#### Scenario: Critical Shot
- **WHEN** a Hunter with Critical Shot I uses it on a creature that is not in combat
- **THEN** the draw is longer than a normal shot and the hit does more damage, by the rule book's numbers

#### Scenario: Rapid Fire
- **WHEN** a Hunter with Rapid Fire turns it on
- **THEN** each draw is shorter and each hit does less damage, by the rule book's numbers
