# Tasks

Browser checks use a scratch server on its own port with a copy of the data. A task is ticked only after its test and its browser check pass.

## 1. Quick wins: pet log and one cast per press

- [ ] 1.1 Pet log. In `effects.js`, `petLine` covers the pet's hit, miss, spell and resist, and blows and spells on the pet, for the pet and its minions by the ids in the `pet` message. `handleEvents` calls it for events that aren't the player's. Verify with `effects.test.mjs` cases (the pet hits for 23, a creature hits the pet for 12, another player's pet logs nothing) and in the browser: a Bonedancer's commander fights a creature and both directions appear in Main.
- [ ] 1.2 One cast per press. Ignore a cast request while casting or within 300 ms of the same request, and ignore key repeats on quickbar keys. Verify with a `spells.test.mjs` case and in the browser: double-click Minor Dusk Dart and see one "You begin casting" and no "already casting" line.

## 2. Client effects through the server

- [ ] 2.1 The importer stores `spell.client_effect`, with the world store migration (a missing column reads 0). Verify with an importer test on a synthetic dump and a world test loading a store without the column. Then re-import and check that Minor Dusk Dart carries 2501.
- [ ] 2.2 `Event.FX` (`fx`) on the `cast`, `spell`, `effect` and `resist` events of player and creature spells. Document it in `docs/protocol.md`. Verify with a world test that the cast and land events of a spell with client effect 2501 carry it, and a protocol test that `fx` is omitted at 0.

## 3. Converting effects and cast animations

- [ ] 3.1 Find what `spells.csv` `Anim1–3` index by reading the client's animation tables in the install. Record the answer in `docs/formats.md`. Verify by naming the three clips Minor Dusk Dart plays for a norseman and a kobold.
- [ ] 3.2 `animconv` writes the `cast_<id>` clips the imported spells use, plus the generic `cast`, for player bases and creature figures. Re-run it over all figures. Verify with an `animconv` test of the role table and by checking `web/data/char/norseman/anim/` holds the cast clips.
- [ ] 3.3 `cmd/effectconv` reads `spells.csv`, `speffects.csv` and `spnifs.csv` from `gamedata.mpk` and the client effects in `world.db`. It converts each needed NIF once (meshes as `propconv` does, emitters via `internal/scenery`), writes `web/data/effects/index.json`, and prints the census of unsupported node types. Document it in `docs/formats.md` and the README run-order. Verify with tests on synthetic tables (effect 2501 → anims 41/42/43, caster 160/161, target 256), and run it over the install with a census of fewer unconverted effects than converted.

## 4. Playing effects in the view

- [ ] 4.1 `web/spellfx.js`: one-shot effect instances attached to a bone, followed each frame, dropped after their duration, capped at 64, and not drawn past the drawing distance. The particles use `particles.js` with a per-instance start and seed, and meshes use the props path. Verify with node tests of the instance lifecycle, the cap and the bone fallback.
- [ ] 4.2 Casts. On `cast`, play the start clip, then hold the loop, and run the caster effects on their bones. On `spell`, play the release. On `interrupt`/`castfail`, clear. This applies to the player's own figure in `viewer.js` and to others in `remote.js`, with the generic `cast` when the clips are missing. Verify with `remote.test.mjs` and viewer tests of the clip sequence, and in the browser: the Bonedancer visibly casts Summon Bone Boss and Minor Dusk Dart, and a creature caster casts too. Save a screenshot for the user.
- [ ] 4.3 Projectiles and target effects. A bolt flies from the caster's hand to the target. The target effect plays on its bone for its duration, and a resisted spell shows no target effect. Arrows fly on the `arrow` event. Verify with node tests of projectile timing, and in the browser: Minor Dusk Dart flies and hits, a heal glows on its target, and a Hunter's arrow flies.
- [ ] 4.4 Fallback. Without `web/data/effects/`, spells cast, land and log as before with no console error. Verify in the browser with the folder moved aside.

## 5. Effects in the HUD

- [ ] 5.1 Pet effects. `protocol.Pet.Fx` is filled in `petMsg` and re-sent when an effect starts or ends on the pet. Document it in `docs/protocol.md`. The pet window shows up to six icons (the least time left first) with time and a name on hover. Verify with a world test (a buff on the pet resends `pet` with `fx`), a `pet` node test, and in the browser: buff the pet and see the icon count down.
- [ ] 5.2 Own effects row. Icons by kind with helpful and harmful frames and the time left, placed under the summary window from its rectangle, replacing the `#effects` badges. Verify with an `effects.test.mjs` case of the frames and order, and in the browser: buff yourself and get slowed, both show, nothing covers them, and the slow's icon goes when the slow ends.

## 6. Close

- [ ] 6.1 Update the specs' scenarios as built, the README (run-order, what's shown), `docs/formats.md`, `docs/protocol.md` and CLAUDE.md's "Open now". Then run `go test -short ./...`, all `web/*.test.mjs` and `openspec validate spell-feedback`. Verify that all of them pass.

## Workflow follow-up

- Archive with `/opsx:archive spell-feedback` once the user has seen the effects in play.
