# Tasks

Each tier ends with a playable server and green tests (`go test -short ./...`, every `web/*.test.mjs`, `openspec validate midgard-playable`). A rule task also updates its spec status, its table test and its registry entry. A protocol task also documents the message in `docs/protocol.md` and adds a protocol or server test.

## 1. Tier 1: staged re-import and the live switch-over

- [x] 1.1 Make `dbimport` write `world.db.next` when `world.db` cannot be replaced, and print that a restart will use it. Verify with an importer test that holds `world.db` open and checks that the staged file is written and the exit code is zero.
- [x] 1.2 At server start, check `world.db.next` (it opens, has the required tables, has spawns), then rotate `world.db` to `world.db.prev` and the staged file into place. A failing staged file is left alone and the reason logged. Verify with server tests for a valid staged store, a truncated one and none, and check that `players.db` is untouched in each.
- [x] 1.3 Document the staged import and the rollback in the README's Multiplayer section. Verify the documented commands run as written on a scratch copy of `server/data/`.
- [x] 1.4 Honour templates marked `ReplaceMobValues` (found during 2.1, design 1a): the spawn takes the template's name, level list, aggro, flags, speed and class type, and the server picks a level from the list at each spawn. Verify with importer tests (a corrected row, a level list, an unmarked template) and a world test of the level pick, and that the creatures near the Galplen start import at level 1 and passive.

## 2. Tier 1: safe start and bind points

- [x] 2.1 Search for a source naming the 1.65 Midgard start town per race and class. Record what was found, or that nothing was, in design.md, and add any source to CLAUDE.md and the README. Verify the decision is written down before 2.2 starts. Found: ZAM's starting-zones table, which agrees with the data (design 2).
- [x] 2.2 Add `rules.UnsafeStart` (aggressive, level 6 or more, within 1,500 units) and the safe-start choice, registered as provisional, or as documented if 2.1 found a source. Verify with the `starting-zone-safety` scenarios as table tests on synthetic spawns.
- [x] 2.3 Apply it at catalog load. `StartFor` returns the replacement, and the load report lists each replaced start with its worst creature and its replacement. Verify with a real-zone test (skipped without `world.db`) that the Troll Warrior start in Galplen is kept on a store imported with templates honoured, that a synthetic unsafe start is replaced, and that the `startup` and `spawn` rows are unchanged.
- [x] 2.4 Add the stored bind point to characters (additive columns), `/bind` at imported bind points, death returning to the bind, and the first-login bind for old characters. Verify with world tests: bind in a town, bind refused in open country, death returns to the bind, and an old character is bound on login.
- [x] 2.5 Move low-level characters found at an unsafe start, once, with a system line. Verify with a world test for a level 1 character (moved) and a level 12 one (not moved).
- [x] 2.6 In the browser, create a Troll Warrior and stand at the start for five minutes unattacked. Look at a level 26 creature from a distance and check it is drawn in purple. Record what was seen in this change folder (no screenshots of game art in git).

## 3. Tier 1: starting equipment and levelling gear

- [x] 3.1 Import `starterequipment` into `start_equip(class, item, ord)`, one row per listed class, with counts, failing when the table is missing or empty. Verify with an importer test on a synthetic dump with a `35;31;22;` row.
- [x] 3.2 Replace `GiveStarterKit` with the table lookup: equip if the slot is free and the class may use the item, otherwise put it in the pack. Missing items are logged, and nothing is chosen any other way. Verify with world tests: the Warrior kit from fixture rows, two weapons for one slot, an unknown item, and an existing character left unchanged.
- [x] 3.3 Add the levelling-gear table to `cmd/parity` (Warrior weapons and armour by five-level band, merchants selling each), with gaps marked. Regenerate `docs/parity` and verify the README there shows the table and that the regression test passes.

## 4. Tier 1: switch the live server

- [ ] 4.1 Ask the user before touching their running server. With their go-ahead:
  - copy `players.db` to a dated backup;
  - run `dbimport` (staged);
  - restart the server on the new build;
  - read the load report.

  Verify that the user's characters load, the Troll start is replaced, and a new character gets the table's kit. Record the steps and the report summary in this change folder.

## 5. Tier 2: combatants and hate

- [x] 5.1 Introduce the combatant interface (player or NPC) and per-NPC hate tables in place of `NPC.target`, with `npcSwing` taking a combatant. Damage adds hate, packmates copy hate, and an NPC fights the highest-hate combatant in its zone. Verify that every existing sim and world test passes unchanged, plus new tests: two players, and the NPC turns to the one who did more damage.
- [x] 5.2 Play a short session in the browser after the refactor: fight, flee past the leash, die, and fight with two clients on one creature. Verify behaviour is as before and record anything that differs.
- [x] 5.3 Check low-level melee against 1.65 (found in 5.2): a level 1 Warrior in the classic kit hits a level 2 creature for 1–2 and takes 10–11, with 39 hit points, and loses. Compare player hit points, weapon damage and creature damage at level 1 with the rule book's sources; fix the rule that is wrong, or record why it is right. Verify with rule-book scenarios at level 1 against a level 1 (yellow) and a level 2 (orange) creature, and a world test of forty fights. Outcome: the rule stands (findings.md 5.3).

## 6. Tier 2: spells and effects

- [x] 6.1 Import the referenced spells, template spell lists, `stylexspell` and item proc columns, with counts and warning totals. Verify with importer tests: a caster template, a missing spell id, and a style-spell row.
- [x] 6.2 Settle the unit of `LifeDrainReturn` from several drain spells in the data, and record it in the `rules-spell-effects` spec. Verify the spec and the 6.3 test agree.
- [x] 6.3 Write `rules/spells.go`: the hit chance, resist and cap, variance, bolt block, damage-over-time ticks, slow factor, mesmerize, stun, drain and heal, each registered with its status. Verify with every `rules-spell-effects` scenario as a table test.
- [x] 6.4 Add effects to combatants, ticked by the zone: start, tick, refresh on recast, end, broken by damage (mesmerize), removed on death and on logout. Send a player their own effects. Verify with world tests: five damage-over-time ticks, a refreshed recast, and a mesmerize broken by damage.
- [x] 6.5 Add the speed factor to `world.Step` and `web/move.js`, regenerate the shared vectors, and send the owner their factor. Stun and mesmerize set it to 0 and block actions. Verify that both movement test suites agree to zero difference with factors 1, 0.6 and 0.
- [x] 6.6 Show effects on the client: icons by effect kind (not by the imported icon number), countdowns, "you are stunned" and similar lines, with a plain fallback. Verify with a node test of the effect view, and in the browser with a slow from a creature.

## 7. Tier 2: creature casting and procs

- [x] 7.1 Write the casting decision (heal below 75%, then a ready hostile spell in range, else melee), the cast state, the failure and interrupt rules, and the 2-second retry. Register them and verify with each `rules-creature-casting` scenario as a table test.
- [x] 7.2 Run casting in the NPC combat step, with `cast` events to nearby clients, and the chat line and cast animation on the client, with no bar for another's cast. Verify with world tests (cast then melee, out of range, interrupted, uninterruptible, healing a packmate) and a protocol test of the event.
- [x] 7.3 Apply style spells, weapon procs and creature offensive procs on landed blows. Verify with each new `rules-melee` scenario, plus a world test where Slam stuns.
- [x] 7.4 Update `cmd/parity` so the implemented creature spell types are counted. Regenerate the baseline, and verify that at least DirectDamage, SpeedDecrease, OffensiveProc, DamageOverTime, Mesmerize, Bolt, Heal, Lifedrain, Stun and DirectDamageNoVariance show as implemented.
- [x] 7.5 In the browser, fight a Midgard caster creature: see the cast line, take a spell, interrupt one, and be slowed. Record what was seen, and add the line-of-sight limitation to the README.

## 8. Tier 2: guards

- [x] 8.1 Classify `GameGuard` NPCs, and make guards attack aggressive creatures of another realm in range, or those fighting their realm's players, chasing to their leash and returning. Verify with world tests for both `server-npcs` guard scenarios.
- [x] 8.2 Count guard damage as an unshared damager in the experience award. Verify with the "A guard does half" scenario as a test.

## 9. Tier 2: groups, group experience and chat

- [x] 9.1 Add groups on `World`: invite by target or `/invite`, accept or decline, a 60-second lapse, leader-only invites, a cap of 8, leave with `/disband`, removal by the leader, leader handover, ending at one member, and leaving on logout. Verify with world tests for every `groups` scenario that does not need the client.
- [x] 9.2 Send each member the group state once a second (name, level, class, health, endurance, zone, in range), to members only. Verify with a protocol test and a world test that a non-member receives none of it.
- [x] 9.3 Write the group experience rules (one damager, split by level among members in range, per-member cap, grey by the highest member in range) and register them as provisional. Verify with each new `rules-experience` scenario as a table test and a world test of a group kill.
- [x] 9.4 Add `/send`, `/tell` and `/g`, with channels on chat lines, the rate limit, and "not in the game". Verify with world tests for each `chat-channels` scenario.
- [x] 9.5 Build the group window from `StatsGroup`/`MiniGroup`, the invitation from `new_group`, and the chat channel colours and Group tab, with plain fallbacks. Verify with node tests of the group view, and in the browser with two clients: invite, accept, see health drop, `/g`, a tell, leave. Repeat with `web/data/ui/` moved aside.

## 10. Tier 2: Warrior abilities

- [x] 10.1 Look for 1.65 sources for Guard's range, Intercept's range and duration, Protect's hate share and Engage's numbers. Record what was found, and add any source to CLAUDE.md and the README. Verify the `rules-abilities` statuses match what was found.
- [x] 10.2 Load abilities from the specialisation-ability rows, and implement Guard, Intercept, Protect and Engage as rules plus world effects. Verify with each `rules-abilities` scenario as a table test and world tests in a group.
- [x] 10.3 Let abilities be used from the skills window and the bar, with refusals as system lines. Verify with the census counting the four abilities as implemented, and in the browser with two grouped clients using Guard.

## 11. Tier 2: horse routes

- [x] 11.1 Spike: find in the data how a ticket names its route (ticket items, `HS_` paths, stable masters' stock lists). Record the link, or its absence and the fallback, in design.md §11 and the README. Verify by listing the Mularn stable master's tickets and the path each resolves to.
- [x] 11.2 Import paths and points. Verify with the "A horse route" importer test.
- [x] 11.3 Add the stable master kind, ticket hand-in, the server-driven ride across zone borders, riders ignored by creatures, and dismount at the end. Verify with world tests for every `horse-routes` scenario.
- [x] 11.4 Show riders on a horse (convert the horse model if the install has one, otherwise draw the rider running). Verify in the browser with two clients that one rides from Mularn and the other sees them pass.

## 12. Tier 2: NPC kinds and the vault

- [x] 12.1 Add the sourced not-classic table, the `notclassic` kind and its answer, and the census column. Verify with the `npc-interaction` scenarios as world tests, and that `cmd/parity` lists teleporters and hasteners as "not classic".
- [x] 12.2 Add the vault keeper kind and vault slots 1000–1039 on inventory rows. Moves are refused out of reach, swaps happen in one transaction, and only the owner is sent the vault. Verify with world and store tests for every `vault-storage` scenario, including a restart.
- [x] 12.3 Build the vault window beside the pack, with a plain fallback. Verify in the browser by storing and taking an item and reopening at another keeper.

## 13. Tier 2: item condition and repair

- [x] 13.1 Add item condition to inventory rows (an additive column, defaulting to 100%), taking it from the template's condition for new items. Verify with a store test that old rows load at 100%.
- [x] 13.2 Write the wear, effectiveness, warning and repair-price rules, registered as provisional except where sourced. Verify with each `rules-item-condition` scenario as a table test.
- [x] 13.3 Apply wear and effectiveness in combat, warn at the thresholds, and add the blacksmith kind with a priced, confirmed repair. Verify with world tests, and in the browser by repairing a worn weapon.

## 14. Tier 2: the game clock and night spawns

- [x] 14.1 Look for the 1.65 day length and night hours. Record the result, and keep 60 minutes and 18:00–06:00 as reference if nothing better is found. Verify the `world-time` statuses match.
- [x] 14.2 Mark night-only spawns at import. Verify with the "A night spawn" importer test.
- [x] 14.3 Add the clock as a function of real time. Night spawns appear over the first game minute of night and leave at dawn, staying while in combat. Add the load count. Verify with world tests using an injected clock: dusk, dawn mid-fight, and a respawn only at night.
- [x] 14.4 Send the clock to clients, and darken the sky and light at night with a fallback. Verify with a node test of the clock view, and in the browser across a dusk.

## 15. Tier 3: character windows and the menu bar

- [x] 15.1 Build ATTRIB. from `StatsAttributes`, and SKILLS and STYLES from `Stats_Spec_Abil`, filled from the owner's sheet data with empty fields left empty. Styles can be dragged to the bar. Verify with node tests of each view, and in the browser that each opens and closes from the menu bar and `C`.
- [x] 15.2 Enable INV., ATTRIB., SKILLS, STYLES and GROUP, and give SPELLS, JOURNAL, MAP and WARMAP a tooltip saying why they are greyed. Verify in the browser, and with `web/data/ui/` moved aside that the plain sheet and panels work.

## 16. Tier 3: orientation, the clock face, proofs and icons

- [x] 16.1 Check the compass against a zone exit of known direction, and fix its north if wrong, using one north for every zone. Verify the `world-orientation` scenarios in the browser, including across a border.
- [x] 16.2 Show the server's game time on the stock clock. Verify the `hud-chrome` clock scenario in the browser.
- [x] 16.3 Watch a real kill's loot through the stock loot window (one item, then Take all), and a real death through the Release window (release early once, wait once). Record what was seen in this change folder.
- [x] 16.4 Do one bounded search for a source that maps the imported icon numbers to this install's sheets, trying it on `Two Fists` (105, sheet 636) and `Bloodletter` (60, sheet 505). Adopt it if it works; otherwise record that none was found.

## 17. Tier 3: carried over from the closed changes

- [x] 17.1 NPC faces (from `npc-equipment` 4.5): find why the converted head texture is `hair_*.png` and fix the skin mapping in the converter. Verify by looking at a guard, a smith and a monster in the browser, and, with the `web/data/` lookup moved aside, that NPCs still draw plain.
- [ ] 17.2 Race and gear checks (from `npc-gear-mapping` 7.3, 7.4 and 7.6):
  - two clients of different races see each other's race and worn gear;
  - a same-race gender pair looks different;
  - equipping changes the look live;
  - in each realm, three armed NPCs, a full-set NPC, a torso-only NPC and a robed caster look right.

  Record the approximations seen in this change folder.
- [x] 17.3 Interior fit (from `jordheim-and-midgard-dungeons` 2.2): run the scored transform search on all 13 Midgard interiors. Verify a test checks that at least 90% of entrance centres and lights fall inside placed pieces, improving Trollheim, Tuscaran Glacier and Nisse's Lair, and record each fit in the README.
- [ ] 17.4 Interior tour (from `jordheim-and-midgard-dungeons` 6.1): for all 13 Midgard interiors, check the spawn view, a walk along a passage and a free-fly overview, and note the frame rate in Abandoned Mines. Record any zone that still looks wrong as a finding.

## 18. Integration

- [ ] 18.1 Play with two clients for an evening's worth of the 1.65 loop: create Warriors, start safely, group, level in Vale of Mularn, ride to another town, bind, bank, repair, see night fall, fight casters, die and pray. Record what broke or felt unlike 1.65 as findings for the next change.
- [ ] 18.2 Regenerate `docs/parity`. Check that rule statuses are listed with no regressions, and update the README's limitations and CLAUDE.md for what changed. Verify all Go and JS tests, `openspec validate midgard-playable --strict`, and the parity regression test pass.
- [ ] 18.3 Switch the user's live server to the final build as in 4.1, with their go-ahead and a fresh `players.db` backup.

## Workflow follow-up

- Commit after each tier, when the user asks.
- Archive with `/opsx:archive` once the user has played it. Unfinished findings from 18.1 go into the next change (the other Midgard classes and player spellcasting).
