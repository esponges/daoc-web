# npc-interaction Specification

## Purpose
TBD - created by archiving change merchants. Update Purpose after archive.

## Requirements

### Requirement: The server knows what kind of NPC it is
The server SHALL classify each NPC by the class type imported with its spawn:
- `GameMerchant` makes it a merchant;
- `GameHealer` makes it a healer;
- a class type naming a class trainer (for example `DOL.GS.Trainer.WarriorTrainer`) makes it a trainer for that class;
- any other class type makes it a kind with no interaction.

An NPC spawned from a world store that lacks class types SHALL have no interaction.

#### Scenario: A merchant spawn
- **WHEN** a spawn imported with class type `GameMerchant` is loaded into a zone
- **THEN** its NPC is a merchant and its entity carries an interaction kind of `merchant`

#### Scenario: A trainer spawn
- **WHEN** a spawn imported with class type `DOL.GS.Trainer.WarriorTrainer` is loaded
- **THEN** its NPC is a Warrior trainer and its entity carries an interaction kind of `trainer`

#### Scenario: A healer spawn
- **WHEN** a spawn imported with class type `DOL.GS.GameHealer` is loaded
- **THEN** its entity carries an interaction kind of `healer`

#### Scenario: A plain creature or a guard
- **WHEN** a spawn has any other class type
- **THEN** its NPC has no interaction kind and an `interact` request for it is answered with a short refusal, not silence

#### Scenario: An old world store
- **WHEN** `world.db` was imported before class types were kept
- **THEN** the server starts, no NPC is interactable, and the log says the store should be re-imported

### Requirement: Interaction is requested by NPC and decided by the server
A client SHALL ask to interact by sending `interact` with an NPC's id. The server SHALL accept the request only when the NPC is in the player's zone, is alive, has an interaction kind, and is within interact range of the player. It SHALL otherwise answer with an error the player can read. A merchant SHALL answer with its shop, a trainer with the player's training state, and a healer with its offer.

#### Scenario: In reach
- **WHEN** a living player stands within interact range of a merchant and sends `interact`
- **THEN** the server answers with that merchant's `shop` message

#### Scenario: A trainer in reach
- **WHEN** a Warrior interacts with a Warrior trainer in reach
- **THEN** the server answers with a `trainer` message listing their specs, levels, next costs and unspent points

#### Scenario: Too far away
- **WHEN** the player is beyond interact range
- **THEN** the server answers with an `err` whose text says the NPC is too far away, and sends no `shop`

#### Scenario: Dead player or dead NPC
- **WHEN** the player is dead, or the NPC is dead
- **THEN** the request is refused and nothing opens

#### Scenario: Unknown or other-zone id
- **WHEN** the id names no NPC in the player's zone
- **THEN** the request is refused without disclosing whether the id exists elsewhere

### Requirement: Right-click interacts and right-drag still turns the camera
The client SHALL treat a right-button press and release that moves less than the existing click threshold over an NPC as an `interact` request for that NPC, SHALL also select it as the target, and SHALL leave a right-drag as camera control only.

#### Scenario: Click on a merchant
- **WHEN** the player right-clicks a merchant without dragging
- **THEN** the merchant is targeted and an `interact` request is sent

#### Scenario: Drag
- **WHEN** the player holds the right button and drags
- **THEN** the camera turns and no request is sent

#### Scenario: Click on nothing, or on a creature with no interaction
- **WHEN** a right-click lands on empty ground or on an NPC with no interaction kind
- **THEN** no `interact` request is sent; the second case only targets it

#### Scenario: Offline viewer
- **WHEN** the viewer runs without a server
- **THEN** right-click behaves as before and nothing is sent
