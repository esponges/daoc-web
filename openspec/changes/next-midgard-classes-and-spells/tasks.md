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
- [x] 3.4 A Troll Berserker was made, levelled to 10, trained Axe to 4 at Krek (the Viking trainer) and two axe styles appeared on the bar, in the browser; the kill itself is `TestMeleeClassesFightWithStyles` (Berserker and Savage, real data), because the browser could not reach a creature at its frame rate.

## 4. Power and casting

- [x] 4.1 No source gave the power pool or regeneration numbers (two searches, see findings), so `rules/power.go` is provisional with table tests for the `rules-power` scenarios and a registry entry.
- [x] 4.2 Power is on the sheet (`mpw`), the snapshot (`pw`, `mpw`) and the HUD (stock summary power bar and a built-in bar for casters); `TestPlayerPower`.
- [x] 4.3 The import now carries the spell lines of the Midgard classes (`line_spell`) and each spell's power; the spellbook is built from the specs a class trains (shared lines plus lines made for the class), sent as `book`, with a "You learn ..." line for new spells.
- [x] 4.4 `cast` request: checks (known, supported, ready, target, power, range), cast time, interrupts by damage and by moving, recast and power spent on landing (`internal/server/world/pcasting.go`, tests in `pcasting_test.go`).
- [x] 4.5 First slice of spell types applied for players: the creature engine's (direct damage, bolt, heal, damage over time, slow, mesmerize, stun, life drain) plus stat and armour buffs and damage add; unsupported types stay in the book marked and refuse a cast. `cmd/parity` counts class spell types by what a player can cast.
- [x] 4.6 SPELLS window from the stock `Stats_Spec_Abil` layout (the menu button is on), a spell bar (Shift+1 to Shift+8, kept in the browser per character), a cast bar and the power bar; without `web/data/ui/` a plain panel (P) does the same. `web/spells.js`, `web/spells.test.mjs`.
- [x] 4.7 Cast a heal, a bolt and the range refusal in the browser (see findings); buffs and a group spell are covered by `TestBuffsRaiseTheSheetAndExpire` and `TestGroupSpellReachesMembers`.

## 5. Left from the last change

- [x] 5.1 The vault window opened at Idonna, the Galplen vault keeper, for the Berserker; repair at a blacksmith and the vault were watched in the last change (`midgard-playable` findings) and the code is unchanged.
- [x] 5.2 A svartalf chanter (Gleam Blast) of the real data casts at a Healer: `TestRealCasterCreatureHurts`. In the browser the camp (level 14 hunters with the chanter) killed the level 10 Berserker in seconds (79 damage a blow), who came back at the bind; the chanter's cast was not isolated.

## 6. Integration

- [ ] 6.1 Two clients play a Warrior, a Healer and a Runemaster in a group through the Vale of Mularn: heal, buff, kill, loot, die, pray. Record findings for the next change (the remaining Midgard classes).
- [ ] 6.2 Regenerate `docs/parity`, update the README and CLAUDE.md, run all Go and JS tests, `openspec validate next-midgard-classes-and-spells --strict`.
- [ ] 6.3 Switch the live server to the new build with a fresh `players.db` backup and the user's go-ahead.
