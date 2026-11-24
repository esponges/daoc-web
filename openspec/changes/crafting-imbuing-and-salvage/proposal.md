# Proposal: crafting-imbuing-and-salvage

## Why

`crafting` made recipes work: a profession, the window, timed crafts, quality, skill gain. Three parts of 1.65 crafting were left out because they need data or sources the first change did not have:

- **Spellcrafting.** Gems are not in the recipe data (no recipes of skill 13 or gem cutting), so imbuing items needs its own source and tables.
- **Stations.** A forge, lathe or alchemy table has to be near; the figures are in the zone placements but were not listed.
- **Salvage.** Returning materials from an item needs a source for the share.

## What Changes

- Station figures recognised in the zone data, and crafting refused away from the one a skill needs.
- Spellcrafting: gems and imbue points, imbuing crafted items, overcharge.
- Salvage of a crafted item at its station.
- Alchemy procs on crafted items.

## Capabilities

### New Capabilities
- `crafting` additions: stations, imbuing, salvage, procs.

### Modified Capabilities
- (none)

## Impact

Server `world/spellcraft.go`, rules additions with sources and registry entries, store columns for an item's imbue record, window additions.
