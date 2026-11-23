# Proposal: crafting

## Why

Crafting was one of the two long-term pursuits of 1.65, alongside RvR, and none of it exists here.

- **The data is already there.** The community database carries the recipes:
  - `crafteditem` gives each recipe's product, tradeskill and level;
  - `craftedxitem` gives its materials.
- **The masters are already placed.** Midgard places Weaponcrafting, Armorcrafting, Fletching, Tailoring, Spellcrafting and Alchemy masters, and the census lists them as not implemented.
- **The windows are already in the install.** The stock layout has `Craft`, `MiniCraft` and `Tradeskills`.
- **Merchant gear runs out at level 30.** Above that, merchant gear is the project's own added stock. Crafted gear is how a 1.65 player outfitted their character past the merchants.

## What Changes

- **Tradeskills.** A character joins one primary crafting profession at its master and gains those skills:
  - Weaponcrafter, Armorcrafter, Fletcher, Tailor, Spellcrafter or Alchemist;
  - the secondary skills (smithing, cloth, leather, wood, metal, gem and herbal work) cap at a share of the primary.
- **Recipes:** from the database, filtered to the 1.65 set.
  - The Craft window lists the recipes each skill can make.
  - Materials are bought from supply merchants and come from the database's merchant lists.
- **Making an item:**
  - a timed craft with a progress bar that moving cancels;
  - the chance of success and of skill gain follow the recipe's level against the crafter's skill (rule book `rules-crafting`);
  - quality from 94 to 100 percent, with masterpieces;
  - the crafter's name goes on the item;
  - a failure loses a share of the materials.
- **Spellcraft:** gems from the Spellcrafter's recipes are imbued into crafted armour and weapons up to the item's imbue points. Overcharge carries a chance to explode.
- **Alchemy:**
  - potions;
  - poisons for the Shadowblade, which joins the existing poison rules;
  - charged and reactive procs on crafted items.
- **Item quality and condition** use the existing condition rules. Quality already matters there, and crafted quality now feeds in.
- **Salvage:** salvaging an item returns some of its materials.
- **Left out:** player housing and the crafting tools that came with it, and the later ToA crafting changes.

## Capabilities

### New Capabilities
- `crafting`: professions, recipes, materials, making an item, quality, spellcraft imbuing, alchemy and salvage.
- `rules-crafting`: success and skill-gain chances, craft time, quality chances, imbue points, overcharge and salvage returns.

### Modified Capabilities
- `merchants`: the supply merchants sell crafting materials.

## Impact

- **Import:** `internal/importer` reads the `crafteditem` and `craftedxitem` tables, and the masters and supply merchants. `world.db` is re-imported.
- **Data:** `players.db` gains a crafting skill table.
- **Server:** new `world/crafting.go`, `world/spellcraft.go` and `world/alchemy.go`, plus rules with table tests and registry entries.
- **Protocol (additive):** recipe lists, craft start, progress and result.
- **Client:**
  - the stock `Craft`, `MiniCraft` and `Tradeskills` windows, with fallbacks;
  - the crafting animation, if the install has one (task 1.3).
- **Census:** a crafting row (recipes implemented of total).
