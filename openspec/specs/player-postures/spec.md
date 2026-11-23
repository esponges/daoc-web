# player-postures Specification

## Purpose
A player's body is in one posture at a time: standing, seated, in the air or swimming, and may sprint. The server decides it and every nearby client draws it with the game's own clips.

## Requirements

### Requirement: Sitting is a server-owned posture
A living player SHALL sit when they ask (`/sit`, the sit key, or the stock SIT button) and stand when they ask again or with `/stand`. The server SHALL keep the posture and send it on the player's entity so every nearby client draws the same thing.

#### Scenario: Sit and stand
- **WHEN** a standing player out of combat sends `/sit`, then `/sit` again
- **THEN** their entity is seated, the log says "You sit down.", then it is standing and the log says "You stand up."

#### Scenario: Others see it
- **WHEN** a player sits within view of another player
- **THEN** the other client plays the sit-down clip, then the seated idle, until the player stands, when it plays the stand-up clip

### Requirement: What stands a seated player up
A seated player SHALL be stood up by moving, jumping, attacking, starting a cast, drawing a bow, using a style, or being attacked. A dead player SHALL not be seated. Sitting SHALL be refused while in the air or swimming.

#### Scenario: Moving stands you up
- **WHEN** a seated player presses forward
- **THEN** they stand up and move, and the client predicts the same

#### Scenario: Attacked while seated
- **WHEN** a creature swings at a seated player
- **THEN** the player is standing on the next snapshot

#### Scenario: No sitting in water
- **WHEN** a swimming player sends `/sit`
- **THEN** they stay swimming and are told they cannot sit here

### Requirement: Jumping
A player on the ground SHALL jump when the jump input is held, rising and falling under one gravity that the server and client share. A jump SHALL not be possible in the air, swimming, seated, or held by a root or stun.

#### Scenario: A jump on flat ground
- **WHEN** a player on flat ground jumps without moving
- **THEN** they leave the ground, return to the same spot, and the server and client agree on every step

#### Scenario: Others see the jump
- **WHEN** a player jumps near another player
- **THEN** the other client plays the jump clip while they rise and the land clip when they touch down

### Requirement: Falling follows the ground
A body whose ground drops away faster than a step can follow SHALL fall under gravity instead of snapping down. It SHALL land when it meets the ground, and the fall height SHALL be passed to `rules-falling`.

#### Scenario: Off a ledge
- **WHEN** a player runs off a ledge ten times their height
- **THEN** they fall over several steps rather than one, land, and take falling damage

### Requirement: Swimming
A body whose feet are below a water surface SHALL swim: it SHALL be held at the surface, move at the swim speed in `rules-movement`, and play the swim clip. It SHALL leave the water when the ground rises above the surface.

#### Scenario: Into a lake
- **WHEN** a player walks into a lake deeper than they are tall
- **THEN** they float at the surface, the posture is swimming, and other clients draw the swim clip

#### Scenario: Out of the water
- **WHEN** a swimming player reaches the shore
- **THEN** they stand on the ground and walk or run again

### Requirement: Sprint
A player SHALL toggle sprint with `/sprint`, the stock SPRINT button or the sprint key. While sprinting and running forward, they SHALL move at the sprint speed and spend endurance as `rules-movement` gives. Sprint SHALL turn off when endurance runs out, and SHALL not work seated, swimming or dead.

#### Scenario: Sprint until tired
- **WHEN** a player with full endurance sprints forward without stopping
- **THEN** they move faster than a run, their endurance falls, and sprint turns off with a message when it reaches zero

#### Scenario: Prediction agrees
- **WHEN** a sprinting player's client predicts their steps
- **THEN** it uses the speed the server sent and is not corrected on flat ground

### Requirement: Posture fallback
A figure converted without a posture clip SHALL fall back to its idle or walk clip rather than fail. The player SHALL still sit, jump, swim and sprint by the rules.

#### Scenario: An old conversion
- **WHEN** a figure with no sit clip sits
- **THEN** it is drawn idle and nothing errors
