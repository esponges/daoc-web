# Tasks

## 1. Rule book and sources

- [x] 1.1 Add the rule book convention to CLAUDE.md ("Where things come from" and "Working here"): one `rules-*` spec per mechanic, statuses, scenarios as vectors, sources listed. Add the same to the README with the sources list from design §12. Verify both files name every source the specs cite.
- [x] 1.2 Create `internal/server/rules` with the registry type: id, capability, requirement, status, sources, scenario names. Add a test that every registered scenario name has a matching subtest. Verify `go test ./internal/server/rules` passes with an empty registry.

## 2. Races, classes and the stat sheet

- [x] 2.1 Write the race table (the five Midgard 1.65 races) and the Warrior class definition, with its eligible races, in the rules package. Verify with the `rules-character-stats` scenarios "The Midgard races" and "Warrior definition" as table tests.
- [x] 2.2 Implement the creation build: point costs, the 30-point budget, the default build, refusals. Verify with "Cost of a stat", "Default build" and "Overspent build".
- [x] 2.3 Implement level stat gains, item stat caps, the 1–350 clamp, the hit point formula with its item cap, and endurance. Verify with "Troll Warrior at 50", "Level 7", "Over the cap", the four hit point scenarios and "Item hit points".
- [x] 2.4 Implement health and endurance regeneration as rules, marked provisional. Verify with "Level 20 resting" and "After a style chain".
- [x] 2.5 Register every rule in groups 2 and 3 with its status and sources. Verify the registry test passes.

## 3. Specialisation and abilities

- [x] 3.1 Implement spec points earned by level and half-level, training costs and the level cap. Verify with every numbered scenario in `rules-specialization`, including "Total at 50" (2,979) and the migration figure (404 at level 20).
- [x] 3.2 Implement item skill bonuses with the 1 + level/5 cap, separate from trained levels. Verify with "A +3 sword at level 10".
- [x] 3.3 Implement ability resolution from a list of (source, level, ability, ability level) rows: highest level wins, and the implemented set is weaponry, armour level, shield size and Evade. Add the object-type-to-ability table. Verify with "Chain at level 10", "Evade at level 15" and "An ability not yet implemented" over fixture rows written in the test (not copied from the dump).

## 4. Melee and creatures

- [x] 4.1 Define the combatant description and the `Rolls` source, and implement the swing order. Verify with "A parried swing" using fixed rolls.
- [x] 4.2 Implement evade, parry and block, with the 120° cone, the attacker divisors, the two-handed halving, and quality and condition on shields. Verify with each defence scenario in `rules-melee`.
- [x] 4.3 Implement the miss chance, its bonuses and the level-capped item bonuses. Verify with "Even fight", "Higher creature", "A style with a to-hit bonus" and "An enchanted weapon".
- [x] 4.4 Implement hit location, the player armour factor and absorption by location, base damage with the DPS cap and the slow-weapon and two-handed factors, the damage cap, weapon skill and variance, resists, truncation, critical hits, style growth damage with its cap and minimum, endurance cost by speed with the half cost on failure, and swing speed. Verify with every damage, style and speed scenario in `rules-melee`.
- [x] 4.5 Implement creature stats, hit points, damage, armour, absorption, swing speed and race resists. Verify with every scenario in `rules-creature-combat`.
- [x] 4.6 Register the melee and creature rules. Add `killMoney` as a provisional rule. Verify the registry test passes.

## 5. Experience and death rules

- [x] 5.1 Write the experience table, the creature value table (with its three departures from the curve), shares, the 1.25 cap and the grey check. Verify with every scenario in `rules-experience` except the protocol one.
- [x] 5.2 Implement con colours by band. Verify with "The bands", "Level 40" and "Level 50".
- [x] 5.3 Implement death and release experience loss with the floor, the death count, constitution loss, and the healer price table and restore amount. Verify with every numbered scenario in `rules-death`.
- [x] 5.4 Register the experience and death rules. Verify the registry test passes.

## 6. World store import

- [x] 6.1 Extend `internal/importer` and its schema with item `bonus`, `condition`, `max_condition` and `durability`, an `item_bonus` table, spawn `race`, race resists, and `spec_ability` from `specxability`. Add classification of trainer class types (`DOL.GS.Trainer.<Class>Trainer`) and healers (`GameHealer`). Verify with importer tests on small hand-written dump fixtures, and that `go run ./cmd/dbimport` reports the new tables' row counts.
- [x] 6.2 Load the new columns and tables in `world.LoadCatalog`, falling back with a log line when the store predates them. Verify with a catalog test on an old-schema store and a new one.

## 7. Player store and migration

- [x] 7.1 Add the character columns and the `char_spec` table as additive migrations in `store.Open`, with load and save of specs, constitution lost, death count, the gravestone and the experience scheme. Verify with a store test that opens an old-schema database (as `store_test.go` already does for gender) and round-trips the new fields.
- [x] 7.2 Convert scheme-0 characters on first join: fill missing base stats with the default build, and move experience to the same fraction of their level. Verify with a world test for "Halfway through 10" and an existing level 20 Warrior getting 404 unspent points.
- [x] 7.3 Let `POST /api/characters` accept an optional build, validated by the rules, and use the default without one. Verify with a gameserver test for a valid build, an overspent one, and none.

## 8. The server uses the rules

- [x] 8.1 Replace `recalcGear` with a sheet recompute on join, equip, level, train and constitution change. Send `stats` when it changes, and add the message to `internal/server/protocol` and `docs/protocol.md`. Verify with a protocol test and a world test for "Equipping a strength item" (a helm with +4 strength).
- [x] 8.2 Replace `canUse` with the ability-based rules: two-handed weapons (moving the shield to the pack, or refusing if it is full), shield size, armour level, level requirement. Fix the `Hand` reading. Verify with the five `character-progression` scenarios as world tests.
- [x] 8.3 Route player and creature swings through the rules resolver. Build creatures from the creature rules at spawn, count attackers, use the 120° cone, charge style endurance on resolution, and take swing times from the rules. Delete the invented functions from `rules.go`. Verify with `sim_test.go` updated to the new numbers, a world test for "A style that is parried", and `go test ./internal/server/...`.
- [x] 8.4 Build the style bar from trained specs. Verify with "Untrained hammer" and "Training opens a style" as world tests.
- [x] 8.5 Award kill experience through the rules (shares, cap, grey), with half-level spec points and level announcements listing the gains. Verify with world tests for "An even solo kill", "Shared", "A grey kill" and the `character-progression` "Level up".
- [x] 8.6 Add `cc` to creature entities per receiving player and document it. Verify with a protocol test for "Target window".

## 9. Trainers, healers, death and praying in the world

- [x] 9.1 Answer `interact` on trainers with `trainer`, and handle `train` atomically with the save. Add both to the protocol and its docs. Verify with world tests for "At the Warrior trainer", "Away from a trainer" and "Another class's trainer", and a protocol test.
- [x] 9.2 Apply the death loss at death. On release (request or timer), apply the release loss, the gravestone, the constitution loss, and the return with full health. Emit `xploss` and `conloss`. Verify with world tests for "Killed by wolves" (level 12), "Killed at level 3", "Release after the first death", "A second grave" and "Back at the bind stone".
- [x] 9.3 Show gravestones as `k:"g"` entities, persist them, restore them on server start, and handle `pray` and `/pray` (5 s, cancelled by moving, fighting or being hit, own grave only, within 2,000 units). Verify with world tests for "Pray", "Someone else's grave" and a cancelled prayer, and a restart test that the grave is still there.
- [x] 9.4 Answer `interact` on healers with `healer`, and handle `restore` in one transaction with the money. Verify with world tests for "Restore at level 12", "Not enough money" and "Nothing to restore".

## 10. Client

- [x] 10.1 Handle `stats`, `trainer`, `healer`, the new events, `cc` and gravestone entities in `web/net.js`. Add node tests for the parsing in a new `web/sheet.test.mjs`. Verify `node web/sheet.test.mjs` passes.
- [x] 10.2 Build the built-in character sheet panel (C key), trainer panel and healer dialog in the plain-panel style. Right-click on a trainer or healer opens them, as for merchants. Verify in the browser: train a spec at the Warrior trainer near the Mularn start, and restore constitution at a healer, with screenshots saved with the change.
- [x] 10.3 Draw gravestones as a targetable marker, colour creature names by `cc`, and put death-penalty and training lines in the chat log. Verify in the browser by dying at level 6 or above, releasing, walking back and praying, with screenshots.
- [x] 10.4 Check that the fallback holds with `web/data/ui/` moved away and the offline viewer at `/?zone=105` is unchanged. Verify by loading both.

## 11. Parity census

- [x] 11.1 Write `cmd/parity`, which reads the dump through `importer.ReadTable` and the rules registry, and writes `docs/parity/parity.json` and `README.md` (names, counts and statuses only, sorted). Verify it runs against `server/data/community/` and that two runs give identical bytes.
- [x] 11.2 Add the regression test against the committed baseline: skipped with `-short` or without the dump, failing when an implemented item or a documented rule goes backwards. Verify by temporarily unregistering the parry rule and seeing it fail with the rule named.
- [x] 11.3 Commit-ready baseline: run the census and check that the README reads sensibly (creature spell types 0 of 15, the Warrior's melee rules implemented, statuses counted). Add a "Parity census" section to the README. Verify the section's command runs as written.

## 12. Integration

- [x] 12.1 Re-import, restart, and play a fresh Troll Warrior from level 1 to 6 in the browser: train, use styles, die and pray. Check that hit points (39 at level 1), damage and experience match the rule book's numbers in the combat log. Verify with screenshots saved with the change, and note any rule that felt wrong as a question, not a change.
- [x] 12.2 Run `go test -short ./...`, every `web/*.test.mjs`, `go test ./...` with the dump and install present, and `openspec validate rules-parity-warrior --strict`. Verify all pass.

## Workflow follow-up

- `stock-ui-followups` 3.1 can use the `stats` message for the stock attributes window.
- Next changes in the Midgard plan: NPC behaviour as written (respawn, aggro, roaming, factions, camp bonus), creature magic, town services, character creation with every Midgard class and race, player magic, groups and guilds.
- Archive with `/opsx:archive` once the review is done, so the `rules-*` capabilities move into `openspec/specs/` as the rule book.
- The browser checks (10.2, 10.3, 10.4, 12.1) ran against a separate test server on port 8778, with a newly imported `world.db` and a copy of `players.db`. The running `gameserver` holds `server/data/world.db` open, so the real store is not re-imported yet. To switch over: back up `server/data/players.db`, stop the server, run `go run ./cmd/dbimport`, and start the server again. Characters convert on their first join.
- Screenshots of these checks show the game's art, which the repository never holds. They are kept outside git: the trainer, praying, the healer, the fallback panels, and a level 10 fight.
- Questions the play-through raised, not changes:
  - A level 10 Warrior still in the level 1 starter kit takes close to the damage cap from even-level creatures (48–52 against 217 hit points) and deals 7–8 a swing. That is the rule as written: armour factor is capped by the item, and DPS by level. Should the starter kit scale, or should the merchants be the answer, as on live?
  - The weapon skill on the sheet is the rule's own number (about 21 at level 6). The game's attribute window showed a larger, differently scaled figure. A display formula could be added if wanted.
