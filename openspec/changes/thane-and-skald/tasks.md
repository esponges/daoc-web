## 1. Sources

- [x] 1.1 Clone DOLSharp and OpenDAoC-Core into the scratch directory. Read how they handle each design open question:
  - the Thane and Skald class numbers;
  - area falloff;
  - pulsing spells: how many are on at once, what stops one, power per pulse;
  - speed enhancement: value, slow combination, combat break;
  - the health regeneration buff's unit and combat rule;
  - the resist buff cap.

  Record each in `findings.md` in our own words with project and file, and where the two disagree. Verify that every open question in `design.md` has an entry.
- [x] 1.2 Run one search per question for a published 1.65 account (Camelot Herald, disorder.dk, Grab Bag), including whether "Call of a thousand storms" is a 1.65 ability. Verify that `findings.md` records each query and its result, empty results included, and add any adopted source to CLAUDE.md and the README.
- [x] 1.3 Update the spec deltas where the findings settle a rule left open (song limit, falloff, speed and combat, regeneration in combat). Verify `openspec validate thane-and-skald --strict` passes.

## 2. Classes

- [x] 2.1 Add the Thane (21) and Skald (24) to `rules.Classes` with the findings' numbers, piety and charisma as casting stats, Viking base and the data's races. Add table tests for the `rules-character-stats` scenarios and a registry entry. Verify with `go test ./internal/server/rules`.
- [x] 2.2 Add world tests on the real data:
  - a Kobold Thane is refused and a Kobold Skald is made with the class kit;
  - Thane and Skald trainers teach their class and not a Warrior;
  - a Thane at Stormcalling 5 has the six spells the spec names.

  Verify `go test -short ./internal/server/world` passes (or skips without the store).

## 3. Area spells

- [x] 3.1 Load the spell `radius` into `SpellDef`, and fall back to 0 for a store without the column. Verify with a catalog test.
- [x] 3.2 Make a range-0 hostile spell castable with no target. Make an area spell land on every living hostile creature within the radius of its centre (the caster or the target), never on players, applying the falloff rule from the findings. Add the area rule to `rules/spells.go` with its registry entry and table tests, and add world tests for the Thunder Shout, Mjollnir and no-target scenarios. Verify with `go test -short ./internal/server/...`.

## 4. Pulsing songs and chants

- [x] 4.1 Import `Pulse` and `PulsePower` into the `spell` table (additive columns), and load them with a fallback for an older store. Verify with an importer test and a catalog test.
- [x] 4.2 Add songs:
  - a pulsing spell toggles on and off, and while on, re-applies to the group in range every `Frequency` tenths of a second;
  - the limit and the stopping conditions are those of the findings;
  - power per pulse is spent if the data asks for it.

  Add world tests for the song, leaving range, toggling off and death scenarios. Verify with `go test -short ./internal/server/world`.
- [x] 4.3 Add the protocol fields: the `pulse` flag on book entries, the `song` field on the effects message, and the `speed`, `regen` and `resist` effect kinds. Document them in `docs/protocol.md` and add a protocol test. Verify with `go test ./internal/server/protocol`.

## 5. Speed, regeneration and resist buffs

- [ ] 5.1 Add speed enhancement: a speed factor above 1, combined with slows and broken by combat as the findings say. Apply it in `world/move.go` and `web/move.js` through `Slow`, and add a shared movement vector with a factor of 1.44. Verify that the Go movement tests and `node web/move.test.mjs` both pass the new vector.
- [ ] 5.2 Add the health regeneration buff to the health tick, and the six resist buffs to the resist a spell meets (strongest per type, capped). Add rules and registry entries with table tests for the `rules-spell-effects` scenarios, and world tests for Song of Rest and the heat chant. Verify with `go test -short ./internal/server/...`.
- [ ] 5.3 Mark the eight types as player-castable in `rules.ImplementedPlayerSpell`, and extend the spell data check to them. Verify that `TestSpellAmountsAgreeWithTheTable` reports no spell, and that `cmd/parity` counts class spell types as 21 of 73.

## 6. Client

- [ ] 6.1 Make the client show songs:
  - the spell bar and SPELLS window mark a song that is on;
  - the effects list draws the `speed`, `regen` and `resist` kinds, both in the stock UI and in the plain fallback.

  Add `web/spells.test.mjs` cases, and verify that every `web/*.test.mjs` passes.
- [ ] 6.2 Confirm that the creation page offers Thane and Skald for their races (it lists what the server allows). Verify in the browser: Norseman shows both, and Kobold shows Skald but not Thane.

## 7. Integration

- [ ] 7.1 On a scratch server, play both classes in the browser and check each result against the spec:
  - a Thane at level 10 casts Thunder Shout into a pair of creatures and both take damage;
  - a Skald in a two-client group turns on Song of Travel, both characters run faster without rubber-banding, and the effect lapses on the member who walks away;
  - a Song of Rest raises the health tick.

  Record what was seen in `findings.md`.
- [ ] 7.2 Bring the docs and checks up to date:
  - regenerate `docs/parity`;
  - update the README (class list, counts, "Not yet" text) and CLAUDE.md ("Open now");
  - run all Go and JS tests and `openspec validate thane-and-skald --strict`.

  Verify that everything passes and `go test ./cmd/parity` agrees with the README.
- [ ] 7.3 Switch the live server:
  - back up `players.db`;
  - re-import the world store;
  - rebuild and restart `gameserver`.

  Verify that existing characters load unchanged and that the creation page lists 29 race and class pairs (22 now, plus 3 Thane and 4 Skald).

## Workflow follow-up

- Archive the change with `/opsx:archive thane-and-skald` once every task is done.
