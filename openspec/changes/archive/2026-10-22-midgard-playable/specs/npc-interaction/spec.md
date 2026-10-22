# Spec Delta

## MODIFIED Requirements

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

## ADDED Requirements

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
