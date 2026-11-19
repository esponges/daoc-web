## 1. Part A: sources for the group essentials

- [x] 1.1 Read DOLSharp and OpenDAoC-Core for understanding only, and search for a published 1.65 account, on:
  - line-of-sight checks: when a cast is checked, and the message;
  - the concentration pool by level, concentration range and grace time;
  - resurrection: the experience returned, the sickness and its length, and the time before release is forced;
  - the numbers of the Healer and Shaman buff types being added: constitution, dexterity, acuity, haste, damage shield, endurance and power regeneration, cures and spread heal.

  Record each with project, file or search, and a status in `findings.md`. Add any new source to CLAUDE.md and the README.

## 2. Part A: line of sight

- [x] 2.1 Add the segment query to `world/collision.go`: a walk over the bins the segment crosses, testing blocking triangles, trunks and the heightmap, and passing when the zone has no collision data. Add `rules` entries `los.visible` and `los.cast` with table tests: open ground, a hill, a dungeon wall, a trunk, no data.
- [x] 2.2 Check a player's cast at its start and at landing in `pcasting.go`. Refuse the cast with the message, spend no power, and filter area-spell targets by line of sight from the centre. Add world tests.
- [x] 2.3 In `casting.go`, a creature without line of sight closes in instead of casting. Add a world test with a wall between them.
- [ ] 2.4 Browser check: cast through a Jordheim wall, past a tree and over a hill, and watch a dungeon caster creature come around a corner. Record what was seen.

## 3. Part A: resurrection

- [x] 3.1 Importer: add the `resurrect_health`, `resurrect_mana`, `concentration` and `is_focus` spell columns, additively, with an importer test. Load them in `catalog_spells.go`. A store without them still loads.
- [x] 3.2 Rules: the resurrection experience return, the sickness and the wait before release is forced, each in `rules-death` with registry entries and table tests.
- [x] 3.3 Server: a resurrection spell targets a dead, unreleased player. Add the offer, accept and decline, raising with the spell's health and power percent, sickness as an effect, and an offer that lapses on release. Add the protocol messages `rez-offer` and `rez-answer` to `docs/protocol.md`, with a protocol test.
- [x] 3.4 Client: the resurrection prompt, in a stock dialog and as a plain fallback. Add a node test.

## 4. Part A: concentration and the Healer and Shaman buffs

- [x] 4.1 Rules: the concentration pool by level and the range and grace time (`rules-concentration`), with registry entries and table tests. Add the pool to the class table and the character sheet.
- [x] 4.2 Effects:
  - effects gain a caster and a concentration cost, and concentration effects have no timer;
  - a once-a-second sweep ends them on range and grace time;
  - caster death, zoning, disconnect and cancel end them.

  Add a test that asserts nothing leaks.
- [x] 4.3 Add the buff types:
  - constitution, dexterity and acuity buffs;
  - haste;
  - damage shield;
  - endurance and power regeneration;
  - the four cures;
  - spread heal.

  Each goes in `rules-spell-effects` with tests. Existing stat buffs that carry a concentration cost become concentration buffs.
- [x] 4.4 Protocol `conc` and `conc-cancel`. Client: the concentration list (stock or plain), with the spell bar marking held buffs. Add node tests.
- [ ] 4.5 Browser check with two clients, a Healer and a Warrior:
  - buff, run out of range and watch the buff end;
  - die holding buffs;
  - cancel from the list;
  - resurrect the Warrior after a death.

  Record what was seen.
- [ ] 4.6 Part A close:
  - regenerate `docs/parity` and update the README and CLAUDE.md;
  - run all Go and JS tests and `openspec validate midgard-to-fifty --strict`;
  - commit, then switch the live server (back up `players.db`, re-import, rebuild, restart).

  Verify that existing characters load.

## 5. Part B: levelling gear to 50

- [ ] 5.1 Sources: what merchant gear 1.65 Midgard sold above level 30, where, and at what quality. Record it in `findings.md`.
- [ ] 5.2 `internal/importer/addedstock.go`: the rule that picks items per band (non-unique, no bonuses, merchant quality, Midgard or neutral) and the merchant lists they extend. `dbimport` writes them flagged as added. Add a test that every band from 31 to 50 is covered for every Midgard weapon kind, shield and armour type.
- [ ] 5.3 Census: the levelling gear table counts added stock, and the gap count goes to 0. Update `go test ./cmd/parity`.
- [ ] 5.4 Browser check: buy and wear a level 45 piece in Jordheim. Then do the part close (as in 4.6).

## 6. Part C: debuffs and creature buffs

- [ ] 6.1 Sources for the debuff, disease, nearsight, confusion, amnesia, bladeturn, ablative, absorption, life transfer and damage-with-debuff numbers. Record them in `findings.md`.
- [ ] 6.2 Debuffs as negative buffs (stat, resist, combat speed), plus the disease, nearsight, confusion, amnesia, unbreakable slow, fatigue consumption, damage-with-debuff and damage-with-slow handlers. Add rule tests and world tests on creatures and players.
- [ ] 6.3 Bladeturn, ablative armour, armour absorption, life transfer and mesmerize duration. Add rule and world tests.
- [ ] 6.4 Creatures cast their template's buffs and debuffs. The creature spell types reach 15 of 15 in the census.

## 7. Part C: pets and charm

- [ ] 7.1 Sources: pet level from the spell, follow and leash distances, charm level limits, charm break and its chance, and the Bonedancer minion limit. Record them in `findings.md`.
- [ ] 7.2 Importer: npc template stats, body type, spells, styles and left-hand chance, additively. Add a test that every Midgard summon spell's template resolves.
- [ ] 7.3 A helper lists the models the summon templates name. Convert them with `charconv`, and add them to the README's first-run recipe.
- [ ] 7.4 Server pets:
  - owner, mode and stance;
  - follow, stay and attack;
  - defensive and aggressive reactions;
  - release on owner death, logout, region change and leash;
  - experience, loot and hate mapped to the owner.

  Add world tests for each spec scenario.
- [ ] 7.5 Bonedancer commander and minions, and charm (taking, breaking, turning hostile). Add world tests.
- [ ] 7.6 Protocol `pet` and `pet-cmd`, and the `/pet` chat command. Client: the pet window (stock or plain) and owner labels. Add node tests.

## 8. Part C: Spiritmaster and Bonedancer

- [ ] 8.1 Sources for their class numbers. Add them to the class table with status. They cast from piety, from base class Mystic.
- [ ] 8.2 Add both classes to the creation page and trainers. Check every spell their lines give at levels 1 to 50 against the census, with world tests for a summon, a damage-with-debuff, a bladeturn and a life transfer.
- [ ] 8.3 Browser check: make each class, summon, command the pet against a camp, and kill with the pet. Record what was seen.

## 9. Part C: Hunter and archery

- [ ] 9.1 Sources: bow draw and damage, Critical Shot, Rapid Fire, Penetrating Arrow, the Hunter class numbers, and Camouflage. Record them in `findings.md`.
- [ ] 9.2 Inventory gains the quiver slots, and merchants sell arrows. Add rules for the draw time and ranged damage with table tests. The server draw state, release, arrow use, line of sight at release, parry turned off, and the three abilities. Add world tests.
- [ ] 9.3 Protocol `shoot` and the draw event. Client: the bow attack key, the draw on the cast bar, and an arrow trail if the effect data allows. Add node tests.
- [ ] 9.4 Add the Hunter class (Beastcraft pet and charm, Spear, Sword, Stealth), the creation page and trainers. Browser check: pull with a bow, fight with the pet, and test charm. Record what was seen.

## 10. Part C: Shadowblade, stealth, poisons and the left hand

- [ ] 10.1 Sources: stealth detection range and speed, what breaks stealth, Danger Sense, Detect Hidden, Distraction, Safe Fall, Climb Walls, the poison chance and charges, the left-hand chance and damage, and the Shadowblade class numbers. Record them in `findings.md`.
- [ ] 10.2 Stealth:
  - hiding and breaking;
  - visibility filtering in the broadcast;
  - creature aggro filtered by detection;
  - stealth speed through the speed factor, with a shared Go/JS movement vector;
  - stealth-opener styles.

  Add world and JS tests.
- [ ] 10.3 Poisons: item charges and spell from the importer, applying to a worn weapon, firing through the weapon-proc path, and poison merchants' stock checked or added. Add tests.
- [ ] 10.4 Left hand: the left-hand slot for trained weapons, the dual swing rule with table tests, and Savage hand-to-hand on the same rule. Add world tests.
- [ ] 10.5 Add the Shadowblade class, the creation page and trainers. Client: a translucent self in stealth, others unseen out of range, and the poison shown on the weapon. Browser check with two clients. Record what was seen.
- [ ] 10.6 Part C close (as in 4.6). The census shows all twelve classes, with any spell type or ability still missing listed with a reason.

## 11. Part D: tasks, and data quests marked not classic

- [ ] 11.1 Sources: whether and how 1.65 Midgard offered kill and delivery tasks, who gave them, the limits, the time limit and the reward. Record them in `findings.md`.
- [ ] 11.2 Mark the 81 Midgard XP-item data quests not classic in `notclassic.go`, and update the census quests line and its test.
- [ ] 11.3 Rules: task limits, time limit and reward, with worked examples per level band, registry entries and table tests.
- [ ] 11.4 Server:
  - task givers chosen by the rule;
  - kill target by con band, and delivery target and task item;
  - progress on kill, delivery on give;
  - expiry and abandon;
  - the `task` table in `players.db`, added on start.

  Add world tests for each spec scenario, including a reconnect.
- [ ] 11.5 Protocol `task`, `ask-task` and `abandon-task`. Client: the task window (the journal's layout if the converted UI has one, plain otherwise) and NPC dialogue. Add node tests.
- [ ] 11.6 Browser check: take, finish, abandon and let a task expire. Then do the part close (as in 4.6).

## Workflow follow-up

- Archive the change with `/opsx:archive midgard-to-fifty` once every task is done. If a part is split into its own change during apply, move its tasks there and note it here.
