# Spec Delta

## Purpose

Gives each character a bank vault, opened at any vault keeper of their realm, to store items they are not carrying.

## ADDED Requirements

### Requirement: Each character has a vault
Each character SHALL have a vault of 40 slots, stored with the character and kept across logouts and restarts. The same vault SHALL open at every vault keeper of the character's realm. Status: reference.

#### Scenario: The same vault in two towns
- **WHEN** a player stores a hammer at the vault keeper in Mularn and later opens the vault in Jordheim
- **THEN** the hammer is there

### Requirement: The vault opens at a vault keeper
Interacting with a vault keeper within reach SHALL open the character's vault next to their pack. Walking out of reach SHALL close it. Account vaults SHALL NOT exist, because they are not classic.

#### Scenario: Opening
- **WHEN** a player right-clicks a vault keeper
- **THEN** the vault window opens with 40 slots and their pack opens beside it

#### Scenario: Walking away
- **WHEN** they walk out of reach
- **THEN** the vault window closes and further moves into it are refused

### Requirement: Items move between pack and vault
A player SHALL move items between their pack and their vault by dragging, while the vault is open. Equipped items SHALL be stored only after being taken off. Each move SHALL be saved in one transaction. A move into an occupied slot SHALL swap the two items.

#### Scenario: Store and take
- **WHEN** a player drags a sword from their pack into an empty vault slot, then drags it back
- **THEN** it is in the vault after the first move and in the pack after the second, and a restart between them keeps it in the vault

#### Scenario: An occupied slot
- **WHEN** a player drags an item from the vault onto a pack slot holding another item
- **THEN** the two items swap places

### Requirement: Only the owner sees a vault
A vault's contents SHALL be sent only to its owner.

#### Scenario: Another player nearby
- **WHEN** two players open their vaults at the same keeper
- **THEN** each sees only their own items
