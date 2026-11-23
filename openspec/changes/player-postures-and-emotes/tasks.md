# Tasks

## 1. Sources and numbers

- [x] 1.1 Find a 1.65-era source for the seated regeneration bonus (Camelot Herald / Grab Bag / patch notes before 1.87). Confirm the factor of two in `rules-resting` or keep it provisional and say so. Add the source to CLAUDE.md and the README.
- [x] 1.2 Pin the falling numbers: safe height, damage share per unit above it, cap, and Safe Fall per level. Read the 1.38 patch notes for the shape and OpenDAoC-Core/DOLSharp for the amounts (understand only, copy nothing). Write them into `specs/rules-falling/spec.md` with worked numbers in the scenarios.
- [x] 1.3 Pin the sprint factor, sprint endurance cost per second and swim speed from the reference servers. Write them into `specs/rules-movement/spec.md` with worked numbers.
- [x] 1.4 Settle the emote set: check for a 1.65-era emote list. Keep D6's classic set unless a source adds commands. Record the command→`anims.csv` column table in the design.

## 2. Rules

- [x] 2.1 `internal/server/rules/resting.go`: seated factor applied to health, power and endurance ticks. Table tests for each `rules-resting` scenario and the updated `rules-character-stats` / `rules-power` scenarios. Registry entries.
- [x] 2.2 `internal/server/rules/falling.go`: `FallDamage(height, maxHP, safeFallLevel)` and `SafeFallLevel(stealthSpec)`. Table tests and registry entries.
- [x] 2.3 `internal/server/rules/movement.go`: `SprintFactor`, `SprintCost`, `SwimSpeed`, `Gravity`, `JumpSpeed`, `StepDown`, `SwimDepth`. Table tests and registry entries.
- [x] 2.4 Write the `rules-resting`, `rules-falling` and `rules-movement` statuses and sources into the rule book registry, and check `cmd/parity` lists them.

## 3. The movement step

- [x] 3.1 `world/move.go`: `Body.VZ/Air/PeakZ`, `Keys.Jump/Sprint`, `Ground.WaterAt`; jump, gravity, step-down falls, landing with `Motion.Landed/FallHeight`, swimming at the surface (D2, D3, D4).
- [x] 3.2 `protocol.FlagJump`, `KeysFromRun` decodes it. Write new vectors to `testdata/move_vectors.json`: jump in place, running jump, ledge, slope that must not fall, water in and out, sprint.
- [x] 3.3 `web/move.js` mirrors 3.1 step for step, and `runKeys` decodes the jump flag. `web/net.test.mjs` replays the new vectors.
- [x] 3.4 `WaterAt` for the server's zone ground (from the `waters` boxes, as `IsWet`) and for the viewer's ground (from `man.waters`).
- [x] 3.5 Server: apply fall damage on `Landed` through `rules.FallDamage`, with the log line and a death through the normal death path. No damage landing in water.
- [x] 3.6 Census test (skipped with `-short` or without the install): count the heightmap edges steep enough to make a sprinting body fall, in every converted zone, and fail if more than one percent of a Midgard start zone's is (the first run: all three are under 0.7 percent, so `StepDown` stays 12).

## 4. Posture, sprint and resting on the server

- [x] 4.1 `world/posture.go`: `sit`, `stand`, `standUp(p)` with messages. Refuse sitting dead, in the air or swimming. The commands are `/sit`, `/stand` and `/sprint`; no new request messages.
- [x] 4.2 Call `standUp` from movement or jump input, attack on, cast start, bow draw, style use, and being swung at (D1). A test for each trigger.
- [x] 4.3 Regen in `zone.go` reads the seated factor. Tests: seated vs standing health, power and endurance over a fixed time.
- [x] 4.4 Sprint: toggle, speed through `Keys.Sprint`, endurance drain in the one-second tick only while moving, off at zero with a message, refused seated, swimming or dead. Add `spr` to the `effects` message. Tests.
- [x] 4.5 Entity fields `ps` and `vz` in snapshots. Protocol additions in `protocol.go` and `docs/protocol.md`, plus a protocol test.

## 5. Emotes on the server

- [ ] 5.1 `world/emotes.go`: the classic set table (name → animation column, lines with and without a target). Commands route through `command()`.
- [ ] 5.2 Refusals (dead, in combat, air, swim), stand-up when seated, the shared chat rate limit. The `emote` event goes to viewers, and the lines go on the "emote" channel to the emoter, the target and onlookers in say range. Tests per scenario in the `emotes` spec.

## 6. Client

- [ ] 6.1 Keys: jump (Space), sit toggle, and a free sprint key after checking `ui.js`/`viewer.js` for clashes; document them in the key help. `net.js` sends the jump flag and predicts sprint from `spr`.
- [ ] 6.2 `cmd/animconv`: new table roles `jump`, `land`, `sitdown`, `sit`, `standup`, `swim`, and lazy `em_<name>` clips for the classic set. Tests in `main_test.go` (roles resolved for a player set, emote clips lazy). Re-run over every player figure.
- [ ] 6.3 `viewer.js` (own figure) and `remote.js` (others): the clip priority in D5, the sit-down → seated → stand-up sequence, jump while rising and land on touchdown, swim, and emotes as one-shots cleared by movement, with fallbacks for missing clips. Node tests for the clip choice.
- [ ] 6.4 `effects.js`/chat: emote lines on the emote channel with the stock Emote colour and tab, fall-damage and sit/stand/sprint lines, and the plain fallback colour.
- [ ] 6.5 `chrome.js`: the stock Command window (D8) with SIT, SPRINT, ATTACK, PRAY and RELEASE wired, the rest greyed with reasons, and a toggle key. Plain Sit/Sprint buttons without `web/data/ui/`. Check it in `/uidemo.html`.

## 7. Verification

- [ ] 7.1 `go test -short ./...`, every `web/*.test.mjs`, `openspec validate player-postures-and-emotes`.
- [ ] 7.2 In the browser (restarted server, two characters): check each of these with a GIF:
  - sit, rest and stand, with regen visibly faster;
  - jump on flat ground and off a ledge in Galplen, with fall damage;
  - swim in a lake and climb out;
  - sprint until endurance runs out;
  - `/wave` at the other character, with both logs and the animation;
  - the Command window buttons.
- [ ] 7.3 Fallback: move `web/data/ui/` aside, confirm the plain Sit/Sprint buttons and emote lines still work, and put it back.
- [ ] 7.4 README (controls, the new rule book entries and their sources) and CLAUDE.md (sources, the open/archived change line).

## Workflow follow-up

- Archive with `/opsx:archive` once every task is checked, syncing the five new and five modified specs.
