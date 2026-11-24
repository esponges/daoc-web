# Tasks

## 1. Engine

- [x] 1.1 Quest JSON schema and loader (D1). Data test: schema plus every reference resolves in `world.db`.
- [x] 1.2 Store: quest progress (D2), with a migration.
- [x] 1.3 `world/quests.go`: offer, accept, step advance for each step kind, quest drops (D3), rewards with choice, abandon. Tests per `quests` scenario.
- [x] 1.4 Talk window integration (`npc-interaction`). Protocol additions, `docs/protocol.md` and a protocol test.
- [x] 1.5 Client: the Quest Journal lists quests and steps, accept and decline in talk, reward choice; plain fallbacks.

## 2. Content, wave 1: epics

- [ ] 2.1 (moved to `classic-quests-content`) Collect sources for each Midgard class epic line. Add them to CLAUDE.md and the README.
- [ ] 2.2 (moved to `classic-quests-content`) Write the epic quest files for the four warrior-type classes (Warrior, Berserker, Savage, Thane).
- [ ] 2.3 (moved to `classic-quests-content`) Write the epic quest files for Skald, Hunter and Shadowblade.
- [ ] 2.4 (moved to `classic-quests-content`) Write the epic quest files for Healer, Shaman, Runemaster, Spiritmaster and Bonedancer.
- [ ] 2.5 (moved to `classic-quests-content`) Check the epic armour templates exist (D4). Play one epic through in the browser.

## 3. Content, wave 2: towns

- [ ] 3.1 (moved to `classic-quests-content`) Mularn and Fort Atla quests (levels 1 to 10).
- [ ] 3.2 (moved to `classic-quests-content`) Haggerfel, Huginfell and Audliten quests (10 to 25).
- [ ] 3.3 (moved to `classic-quests-content`) Jordheim and higher-level camp quests.

## 4. Census and docs

- [ ] 4.1 (moved to `classic-quests-content`) `cmd/parity` quests row counts the authored quests. README and CLAUDE.md.

## Outcome (engine done, content just begun)

- Done: 1.1 to 1.5 as a plain panel (the stock journal is not wired), and four quests (`server/quests/midgard/`: Simple Misgivings, Mauler Invasion, Bolli's Leather Supplies, Trapper's Pride; 3.1 in part), checked against the world data by `TestQuestDataResolves`.
- Not done: the epic lines (2.x) and the town quests (3.x). Each needs a public description to retell; the web search found scattered pages, not lists. 4.1 (a census row) waits for more content.
