# Design: classic-quests

## Context

- **Tasks** (`world/tasks.go`) already give kill and delivery tasks from town NPCs, with talk-window integration and journal entries. The quest engine generalises the same pieces.
- **Talk text:** NPC talk text comes from the world data (`npc-interaction`). Quest text is added on top.
- **Licensing:** the world names (NPCs, creatures, items) are already in the imported data. Quest content is our own text, and is committed, unlike game data. It must stay our own wording.

## Decisions

**D1. Quests as JSON in the repo.**
- `server/quests/<realm>/<id>.json` holds:
  - id, title, level range, conditions, source URL;
  - steps: each with a kind, its target references by database ID or name, a count, and journal text;
  - rewards.
- The server loads them at start.
- *Alternative:* store quests in `world.db`. Rejected: they are ours, not imported.

**D2. Progress.** `players.db` gets `char_quest(char_id, quest_id, step, counts_json, done_at)`.

**D3. Quest drops** hook the loot roll: for each player on the killer's credit list with an active step needing the item, roll the step's drop chance.

**D4. The epic armour** comes from the database item templates where they exist (they model 1.65). The census reports any piece missing.

**D5. Waves.** The class epics come first, because they matter most to a level 50 group. Town quests follow, by town. The census counts the authored quests and lists the next ones.

## Risks / Trade-offs

- [Lots of hand-written content] → one quest file per quest, a strict schema test, and waves small enough to play through in the browser.
- [Sources vary in accuracy] → each quest cites one; mismatches are fixed in play.
