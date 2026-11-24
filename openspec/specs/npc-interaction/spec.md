# npc-interaction Specification

## Purpose
TBD - created by archiving change merchants. Update Purpose after archive.

## Requirements

### Requirement: The server knows what kind of NPC it is
The server SHALL classify each NPC by the class type imported with its spawn:
- `GameMerchant` makes it a merchant;
- `GameHealer` makes it a healer;
- a class type naming a class trainer (for example `DOL.GS.Trainer.WarriorTrainer`) makes it a trainer for that class;
- `GameStableMaster` makes it a stable master;
- `GameVaultKeeper` makes it a vault keeper;
- `Blacksmith` makes it a blacksmith;
- `GameGuard` makes it a guard, with no interaction;
- a class type on the not-classic list makes it a not-classic NPC;
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

#### Scenario: Travel and services
- **WHEN** spawns with class types `DOL.GS.GameStableMaster`, `DOL.GS.GameVaultKeeper` and `DOL.GS.Blacksmith` are loaded
- **THEN** their entities carry the interaction kinds `stable`, `vault` and `smith`

#### Scenario: A plain creature or a guard
- **WHEN** a spawn has class type `DOL.GS.GameGuard` or another kind with no interaction
- **THEN** its NPC has no interaction kind, and an `interact` request for it is answered with a short refusal, not silence

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

### Requirement: NPCs that 1.65 did not have say so
The server SHALL hold a list of NPC class types that were not part of 1.65, each with the reason and the source that dates it. A type SHALL leave the list only when a source shows it in 1.65. NPCs of those types SHALL still spawn and be drawn, and SHALL answer an interaction with a line saying their service is not part of this world.

#### Scenario: The starting list
- **WHEN** no source has yet been recorded for them
- **THEN** the list holds the teleporters (`LiveTeleporter`, `OFTeleporter` and their assistants), the hasteners (`GameHastener`, `FrontierHastener`), the account vault keeper, the market explorer, and the seal and bounty merchants

#### Scenario: A teleporter
- **WHEN** a player right-clicks a `DOL.GS.Scripts.LiveTeleporter` in Mularn
- **THEN** the NPC says its service is not part of this world, and the player is not moved

#### Scenario: Counted, not missing
- **WHEN** the parity census runs
- **THEN** these kinds are listed as "not classic" with their reason, not as missing

### Requirement: Quest talk
Talking to an NPC SHALL show, after its own text, the quests it offers to this player and any active quest step that ends with it. Delivering or finishing SHALL happen in the same talk window.

#### Scenario: Turning in
- **WHEN** a player whose current step is "bring the pelt to Svala" talks to Svala with the pelt
- **THEN** the talk window offers to hand it over, and doing so advances the quest
