## 1. Playtest findings

- [x] 1.1 Show the attack state on the quickbar or target window, and stop a repeat press of the attack key from silently cancelling. Verify in the browser.
- [x] 1.2 Look for a source on 1.65's attack toggle and auto-walk to a target. Adopt it (add to CLAUDE.md and README sources) and implement, or record in `findings.md` that none was found and behaviour stays.
- [x] 1.3 Look for a source on night spawns appearing at dusk. Same rule as 1.2.
- [x] 1.4 Pack-to-pack `equip` is the pack's drag-to-rearrange (`web/ui.js` onMove), so it is kept, not rejected; a server test (`TestPackSlotsRearrange`) pins it. Finding 5 was a script mix-up, not a bug.
- [x] 1.5 Loot by creature level (`loot-and-release` delta): no 1.65 source found for the generators, so `rules.GenericDropPercent` (22% at level 1 rising to 40%) is registered as provisional with a table test. Fitting items to the creature kind needs item-kind data the import lacks, so it is left out and the spec says so.

## 2. Gender

- [x] 2.1 Female bases are already converted for all six Midgard races (`races.json`: Norseman, Troll, Dwarf, Kobold, Valkyn, Frostalf), so nothing is missing.
- [x] 2.2 The `gender` column, `fm` on the wire and the male default already existed; creation now takes `female` (docs/protocol.md, `TestCreateCharacterGender`).
- [x] 2.3 Add the choice to the creation page and draw the right model and face on every client, with the male fallback. Verify two clients of the same race and different genders in the browser.

## 3. Classes

- [x] 3.1 Write the class table entries for Berserker, Savage, Healer, Shaman and Runemaster in `rules/character.go` with sources, a rule-book spec update and table tests.
- [x] 3.2 Spec lines, styles, trainers and kits were already imported for every class; `TestOfferedClassesOnRealData` and `TestNewClassesTrainAndGainStyles` check each new class on the real data.
- [x] 3.3 The creation page lists every (race, class) pair the rules allow (24 characters); the server refuses other pairs (`TestCreateCharacterByClass`).
- [ ] 3.4 Level, train and fight a Berserker and a Savage in the browser; record the styles seen.

## 4. Power and casting

- [ ] 4.1 Search for the power pool and regeneration sources; write `rules-power` with status, tests and a registry entry.
- [ ] 4.2 Add power to the sheet, the protocol (additive) and the HUD power bar.
- [ ] 4.3 Build the spellbook from the imported lines by class and level, with level-up additions and messages.
- [ ] 4.4 Add the cast request: checks, cast time, power spend, interrupts and recast, reusing the creature engine. Protocol and server tests.
- [ ] 4.5 Implement the first slice of spell types: direct damage, bolt, heal, damage over time, damage add, stat and armour buffs, speed decrease, mesmerize, stun. Unsupported types are marked and refuse a cast. Update `cmd/parity` to count class spell types.
- [ ] 4.6 Build the SPELLS window from the stock layout, spells on the quickbar and a cast bar, with the built-in panel as the fallback. Check with `web/data/ui/` moved aside.
- [ ] 4.7 Cast a heal, a bolt and a buff in the browser against a real creature and a group member.

## 5. Left from the last change

- [ ] 5.1 Use the vault, repair gear and bank in play and record what was seen.
- [ ] 5.2 Fight a caster creature in play and record what was seen.

## 6. Integration

- [ ] 6.1 Two clients play a Warrior, a Healer and a Runemaster in a group through the Vale of Mularn: heal, buff, kill, loot, die, pray. Record findings for the next change (the remaining Midgard classes).
- [ ] 6.2 Regenerate `docs/parity`, update the README and CLAUDE.md, run all Go and JS tests, `openspec validate next-midgard-classes-and-spells --strict`.
- [ ] 6.3 Switch the live server to the new build with a fresh `players.db` backup and the user's go-ahead.
