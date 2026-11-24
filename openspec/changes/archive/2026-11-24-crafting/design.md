# Design: crafting

## Context

- **Recipes:** in the community data, `crafteditem` (product, `CraftingLevel`, `CraftingSkillType`) and `craftedxitem` (materials and counts) hold every recipe. They are not imported today.
- **Item templates:** already imported, and items have quality and condition (`rules/condition.go`).
- **Masters:** they are placed but skipped (`WeaponCraftingMaster` and the others in the census).

## Goals / Non-Goals

**Goals:**
- A friend can level a profession from 1 to the 1.65 cap and make gear and potions for the group.

**Non-Goals:**
- Housing tools, consignment and the later crafting changes.
- Crafting for the other realms: Albion and Hibernia recipes come with `albion-realm` and `hibernia-realm` through the same tables.

## Decisions

**D1. Filter recipes to 1.65.** The community recipe set includes later items. Recipes whose product template is a later-era item are dropped, by the same not-classic marking the importer already uses (`internal/importer/notclassic.go`). The census counts what remains.

**D2. Stations are props.** Forges, lathes and alchemy tables are recognised by their figure names in the zone placement data. Task 1.2 lists them. A station is "near" within 512 units.

**D3. Skill storage.** `players.db` gets `char_craft(char_id, skill, value)` and a `char.profession` column.

**D4. Item bonuses.** Gems write into the item instance's bonus list, which the character stat code already sums (`character.item-stats`). Instances gain a `crafter` field and an imbue record.

**D5. Rules.** Success, quality, time, skill gain, imbue points and overcharge go in `rules/crafting.go`.
- Each carries sources: Herald crafting pages, 1.65-era crafting guides, and the Grab Bag.
- Where only a reference server shows the number, the status is reference.

## Risks / Trade-offs

- [Large recipe set, thin 1.65 numbers] → keep the curves in one rules file with status, and tune them in play.
