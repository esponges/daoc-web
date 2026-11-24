# Proposal: classic-quests-content

## Why

`classic-quests` built the engine and four quests. The rest of Midgard's quests (the class epic lines with their armour, and the town and camp quests from level 1 to 50) are content, and each needs a public description to retell in our own words.

## What Changes

- Write quest files in `server/quests/` in waves: the class epics first, then the towns by level (Mularn, Fort Atla, Haggerfel, Huginfell, Audliten, Jordheim, camps).
- Each file cites one public source, uses names present in the world data (the data test checks), and pays from the source's figures.
- The census counts authored quests.

## Capabilities

### Modified Capabilities
- `quests`: more content, no new rules.

## Impact

Data files and the parity census only.
