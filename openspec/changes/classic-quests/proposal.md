# Proposal: classic-quests

## Why

1.65 Midgard had quests at every level: town and camp quests, item and kill quests, the level 5 class quests, and each class's epic quest line to level 50 with its epic armour. Here there are only tasks, which end at level 20.

The community database has no classic quests. Its 81 data quests are later "XP Item" collections, which the census marks not classic. So every quest has to be written by this project, in its own words, from public descriptions of what each quest asked: who gives it, where to go, what to kill or fetch, and the reward. The stock layout already has the Quest Journal, built in `stock-windows`.

## What Changes

- **A quest engine.** Quests are data files in the repo, written by this project. Each quest is:
  - steps: talk to an NPC, kill N of a creature, collect a drop, deliver an item, go to a place, or interact with an object;
  - with conditions: level, class, realm, and an earlier quest done;
  - and rewards: experience, money, items, and a choice of reward items.
- **Givers and markers.** NPCs that give a quest show it in their talk text. A quest's targets drop its items only while the quest is active.
- **The journal** shows active and done quests and the current step, and abandoning a quest is allowed.
- **Content, in waves:**
  - first, the Midgard class epic lines (Warrior, Berserker, Savage, Thane, Skald, Healer, Shaman, Runemaster, Spiritmaster, Bonedancer, Hunter, Shadowblade), with their epic armour sets;
  - then the town quests of Mularn, Fort Atla, Haggerfel, Huginfell, Audliten and Jordheim, level by level.
- **Each quest file cites its source:** a public quest walkthrough (Allakhazam or a Herald quest page), retold in our words. NPC names and places are those already in the world data.
- **No copying of text.** Quest dialogue is written fresh. It is not copied from a walkthrough or a GPL server.

## Capabilities

### New Capabilities
- `quests`: the quest engine (steps, conditions, rewards, quest drops, journal, abandon) and how quest content is authored and sourced.

### Modified Capabilities
- `npc-interaction`: talking to a quest giver offers, advances and completes quests.

## Impact

- **Repo:** a new `server/quests/*.json` (our own text, committed), with a schema test.
- **Server:** `world/quests.go`; `players.db` quest progress (a migration); loot generation for quest drops.
- **Protocol (additive):** quest offer, update and journal entries.
- **Client:** the stock Quest Journal shows real quests, and the talk window offers accept and decline.
- **Census:** quests implemented against the authored list.
