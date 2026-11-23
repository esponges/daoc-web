## ADDED Requirements

### Requirement: Choosing a profession
A character SHALL join one primary profession at its master, and SHALL be able to change only by starting over in the new profession. Joining SHALL grant the profession's skills at 1.

#### Scenario: Joining
- **WHEN** a character with no profession asks the Weaponcrafting master to join
- **THEN** they are a Weaponcrafter with weaponcraft and its secondary skills at 1

### Requirement: Recipe list
The Craft window SHALL list each recipe of the character's skills with its level and materials, and mark those the character can make now.

#### Scenario: Low skill
- **WHEN** a Weaponcrafter at 50 opens the Craft window
- **THEN** recipes up to their reach are listed, and those above it are marked too difficult

### Requirement: Making an item
Crafting SHALL require the materials in the bag and a forge, lathe or alchemy table nearby as the recipe's skill needs. It SHALL take the recipe's craft time, and be cancelled by moving, combat or casting. When it ends it SHALL succeed or fail by the rules. Success SHALL consume the materials and give the item with a quality and the crafter's name. Failure SHALL consume a share of the materials.

#### Scenario: Success
- **WHEN** a crafter with enough skill and materials crafts a sword at a forge and succeeds
- **THEN** the materials are gone, the sword is in the bag with its quality, and the crafter may gain a skill point

#### Scenario: No forge
- **WHEN** a crafter starts a metal recipe away from any forge
- **THEN** it is refused with "You must be near a forge to make this."

### Requirement: Skill gain
A craft SHALL raise the recipe's skill by one with the rule book's chance, which depends on how far the recipe's level is from the skill. A recipe far below the skill SHALL give no gain.

#### Scenario: Grey recipe
- **WHEN** a crafter at 300 makes a level 100 recipe
- **THEN** they gain no skill

### Requirement: Spellcraft
A Spellcrafter SHALL imbue crafted items with gems, up to the item's imbue points by level and quality. Each gem SHALL add its bonus to the item. Going past the points SHALL carry the rule book's chance of the item exploding, destroying it and harming the crafter.

#### Scenario: Within points
- **WHEN** a Spellcrafter imbues two gems within a 99 percent item's points
- **THEN** both bonuses are on the item and nothing can explode

### Requirement: Alchemy
An Alchemist SHALL make potions and poisons, and add charges or reactive procs to crafted items. Poisons SHALL use the existing poison rules.

#### Scenario: A poison
- **WHEN** an Alchemist makes a poison and a Shadowblade applies it
- **THEN** it behaves as the merchant-bought poison of the same kind

### Requirement: Salvage
Salvaging an item at the right station SHALL destroy it and return some of its base material, in the rule book's share.

#### Scenario: Salvage
- **WHEN** a crafter salvages a metal sword at a forge
- **THEN** the sword is gone and metal bars are in the bag
