## ADDED Requirements

### Requirement: Server owns world state
The server SHALL hold the authoritative position, health and state of every character and NPC in a zone and SHALL send clients snapshots of the part of the world near them.

#### Scenario: Two players in one zone
- **WHEN** two players are in zone 100 within sight of each other
- **THEN** each client receives the other's position, heading, current animation role and equipment, and draws it

#### Scenario: Out of range
- **WHEN** another player is beyond the interest distance
- **THEN** the client is not sent that player until they come into range

### Requirement: Movement is validated
The server SHALL accept movement input, apply the game's walking, running and strafing speeds on the server's own heightmap, and correct a client whose reported position disagrees.

#### Scenario: Honest movement
- **WHEN** a client runs forward for one second
- **THEN** the server moves the character about 191 units, and the client's predicted position stays within a small tolerance

#### Scenario: Speed hack
- **WHEN** a client reports a position farther than the maximum speed allows
- **THEN** the server rejects it, keeps the character at the last valid position and sends a correction

#### Scenario: Below the ground
- **WHEN** a client reports a position under the terrain
- **THEN** the server places the character on the ground instead

### Requirement: The local player is predicted
The client SHALL apply its own input immediately and reconcile with the server's corrections without visible snapping for errors within tolerance.

#### Scenario: Latency
- **WHEN** the round trip is 100 ms
- **THEN** the local character responds to input on the next frame

### Requirement: Chat
The server SHALL deliver say messages to players near the sender and broadcast messages to all, and SHALL limit the rate and length of messages.

#### Scenario: Say
- **WHEN** a player says "hello"
- **THEN** players within say range see it with the sender's name, and others do not

#### Scenario: Flooding
- **WHEN** a client sends messages faster than the limit
- **THEN** the excess is dropped and the sender is told

### Requirement: Protocol is versioned and tolerant
The client and server SHALL exchange a protocol version on connect and the server SHALL refuse a mismatched client with a readable message, ignoring unknown message types rather than disconnecting.

#### Scenario: Old client
- **WHEN** a client with an older protocol version connects
- **THEN** the server closes with a message asking for a reload

### Requirement: Disconnects are clean
The server SHALL save a character and remove it from other clients' view when its connection closes, and SHALL keep it in the world for a short grace period if it was in combat.

#### Scenario: Close the tab
- **WHEN** a player closes the browser tab
- **THEN** their character is saved and disappears for others after the grace period

### Requirement: Movement is spent against the clock
The server SHALL move a character only by running its movement steps, at a rate of one fixed step per sixtieth of a second of real time, and SHALL NOT take a position from the client.

#### Scenario: Ten seconds sent at once
- **WHEN** a client sends ten seconds of running steps in one message and one second passes
- **THEN** the character has moved about one second's distance, not ten, and a position the client claims to have reached is not believed

#### Scenario: Prediction and server agree
- **WHEN** a client predicts a path with the same steps
- **THEN** its position after the server's acknowledgement differs from the server's by under a unit
