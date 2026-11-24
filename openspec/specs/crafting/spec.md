# crafting Specification

## Purpose
Tradeskills: joining a profession at its master, recipes, timed crafts with quality and a maker's name, skill gain, and stacking materials.

## Requirements

### Requirement: Choosing a profession
A character SHALL join one primary profession at its master (the one it last talked to), and SHALL keep it. Joining SHALL grant the profession's skills at 1.

#### Scenario: Joining
- **WHEN** a character with no profession asks the Weaponcrafting master to join
- **THEN** they are a Weaponcrafter with weaponcrafting and its secondary skills at 1, saved with the character

#### Scenario: One craft
- **WHEN** a crafter asks another master to join
- **THEN** it is refused

### Requirement: Recipe list
The crafting window SHALL list each recipe within reach of the character's skills with its level, chance, time and materials held against needed, and mark those the character can make now.

#### Scenario: Low skill
- **WHEN** a Weaponcrafter at 1 opens the window
- **THEN** only recipes up to 50 levels above their skill are listed

### Requirement: Making an item
Crafting SHALL require the materials in the pack. It SHALL take the recipe's craft time, and be cancelled by moving, combat or casting. When it ends it SHALL succeed or fail by the rules. Success SHALL consume the materials and put the item in the pack with its quality and the crafter's name; the quality SHALL count in combat. Failure SHALL consume half of each material stack.

#### Scenario: Success
- **WHEN** a crafter with enough skill and materials crafts a dagger and succeeds
- **THEN** the materials are gone, the dagger is in the pack with its quality and maker, and the crafter may gain a skill point

#### Scenario: Cancelled
- **WHEN** a crafter moves away or is attacked during a craft
- **THEN** the craft stops and nothing is used up

### Requirement: Alchemy and other recipes
Potions, poisons, arrows and every other product in the recipe data SHALL be made like any item; poisons made so SHALL behave as the bought poison of the same item.

#### Scenario: A poison
- **WHEN** an Alchemist makes a poison
- **THEN** it is the same item a merchant sells, with the maker's name

### Requirement: Supply merchants
Materials SHALL stack in the pack, and buying more of one SHALL add to its stack.

#### Scenario: Buying bars
- **WHEN** a crafter buys a bar while holding a stack of them with room
- **THEN** the stack grows by one and takes no new slot
