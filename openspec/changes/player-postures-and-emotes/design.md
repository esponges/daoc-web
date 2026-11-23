# Design: player-postures-and-emotes

## Context

- **Movement:** it is one fixed 1/60 s step, `Step` in `internal/server/world/move.go`, mirrored in `web/move.js`.
  - Shared vectors in `internal/server/world/testdata/move_vectors.json` keep the two in step.
  - The step only moves in X/Y and then snaps Z with `GroundAt`.
  - The server takes no position from the client, so anything vertical is the server's alone.
  - Input runs are `[count, fwd, strafe, turn, flags]`, and only `FlagRun` (bit 0) is used.
- **Regeneration:** it lives in `zone.go` (endurance every second, power every 3 s, health every 3 s) and reads the `rules` package.
- **Animations:**
  - `cmd/animconv` takes a fixed role list (`tableRoles`) from `anims.csv` and `canims.csv` per anim set, plus cast clips (lazy).
  - Every player anim set (troll set 24 checked) also fills: Jump `ju_h_a`, Land `ju_h_b`, Sit Down / Sit Idle / Stand Up (per race), Crawl (the swim clip `Sw_h`), and the emote columns 53–69 and beyond.
  - None are converted today.
- **Water:** each zone's manifest has `waters` boxes and a surface height. `static.go IsWet` already answers "is this under water" for NPCs.
- **UI data:**
  - The stock layout has a `command` window: ATTACK `ToggleAttackMode`, SPRINT `Sprint`, SIT `ToggleSitting`, FOLLOW, STICK, FACE, PRAY, RELEASE, BONUSES.
  - The chat config has an "Emote" filter.
  - The client's command table in `game.dll` lists the emote commands by name.

## Goals / Non-Goals

**Goals:**
- One movement step, shared by server and client, that jumps, falls, lands and swims.
- Posture is a server state, drawn the same on every client with the game's own clips.
- The rule numbers (resting, falling, sprint, swim) live in the rule book with status and sources.

**Non-Goals:**
- **Out of scope:** the alternative sit poses (squat, knee-up and so on, columns 96–105), mounts beyond the existing horse routes, ladders, boats, and FOLLOW, STICK and FACE. These are later or separate work.
- **Not modelled:** breath or drowning (no such mechanic at 1.65 that we know of) and turning-in-place clips (columns 3–4 point at rows missing from `animnifs.csv`).
- **NPCs:** they do not jump, swim or sit. They keep avoiding water as today.

## Decisions

**D1. Posture is a player field with one stand-up path.**
- The player gets `posture` (standing, seated) and `sprint` (bool).
- One `standUp(p)` is called from:
  - input with any movement or jump;
  - attack on;
  - cast start;
  - bow draw;
  - style use;
  - being swung at (in the combat code where a swing picks its target).
- Air and swim are not stored: they come out of the body's state after each step.
- *Alternative:* the client decides it is seated and the server trusts it. Rejected because regeneration depends on it.

**D2. Vertical motion inside `Step`.**
- `Body` gains `VZ` and `Air`, and `Keys` gains `Jump`.
- The order each step: horizontal move as today, then the vertical part:
  - **Jumping:** on the ground, with `Jump`, not held and not swimming: `VZ = JumpSpeed`, `Air = true`.
  - **In the air:** `VZ -= Gravity*dt` and `Z += VZ*dt`. When Z falls to the ground, land: `Air = false` and `VZ = 0`, and `Motion` reports `Landed` with `FallHeight` (from the highest Z in this airborne spell to the landing Z, carried in `Body.PeakZ`).
  - **On the ground:** if the new ground is more than `StepDown` below Z, go airborne with `VZ = 0`. Otherwise snap as today.
- `StepDown` is 12 units a step. That lets a sprint down a 45° slope (7 units a step) stay grounded, while a ledge falls. The census check in task 3.6 runs every Midgard zone's start path to confirm no false falls.
- Numbers (project, `rules-movement`):
  - `Gravity = 500` u/s² and `JumpSpeed = 200` u/s;
  - that gives a 40-unit apex and 0.8 s of air;
  - it matches the 0.4 s rise clip plus the land clip.
- The vectors gain cases: a jump in place, a running jump, a ledge, a slope that must not fall, and water entry and exit.
- *Alternative:* a client-only cosmetic jump. Rejected because falling damage and ledges need server Z anyway.

**D3. Water through the Ground interface.**
- `Ground` gains `WaterAt(x, y) (surface float64, ok bool)`, built from the same `waters` boxes as `IsWet`. The JS ground gets the same from `man.waters`.
- A body is swimming when `surface - ground > SwimDepth`. `SwimDepth` is 0.7 of a human figure's height and is a project constant.
- A swimming body is held at `surface - SwimDepth` and moves at `SwimSpeed` (`rules-movement`).
- Falling into water lands at the surface with no damage.

**D4. Speed as a multiplier, sprint included.**
- `Keys.Slow` stays as is.
- Sprint is a separate `Keys.Sprint` that multiplies run speed by `SprintFactor` only when running forward and not swimming.
- The server sets it from `p.sprint`, and the client gets `sprint` in the existing `effects` message's speed fields, so prediction matches.
- Endurance drain sits in the zone's one-second tick, charged only for seconds the body moved.
- *Alternative:* fold sprint into the slow factor. Rejected because a slow and a sprint stack differently.

**D5. Clips: new table roles, one priority order.**
- `tableRoles` gains `jump` (Jump), `land` (Land), `sitdown` (Sit Down), `sit` (Sit Idle), `standup` (Stand Up) and `swim` (Crawl).
- Each classic emote becomes `em_<name>` from its column, written lazy as cast clips are.
- In `viewer.js` (own figure) and `remote.js` (others), clips are picked in this order:
  1. dead;
  2. air (jump while rising, then held; land on touchdown);
  3. swim;
  4. seated (sit-down once, then the seated idle; stand-up once on leaving);
  5. locomotion as today.
- An emote is a one-shot action over the idle, cleared by any movement.
- A missing clip falls back to idle or walk.

**D6. The classic emote set.**
- The 1.65 set is the original emote columns 53–69 of `anims.csv` and the commands that name them in the client:
  - beckon, blush, bow, cheer, clap, cry, curtsey, dance, flex;
  - kiss, laugh, point, salute, victory, wave;
  - bang (the shield bang).
- Later columns are left out unless task 1.4 finds a 1.65-era source listing them:
  - Korean (ponder, military, present);
  - taunt, rude, hug, letsgo, meditate, no, raise, shrug, slap, slit, surrender, yes;
  - beg, induct, dismiss, knock;
  - the "Classic_em_*" set: yawn, rofl, …
- The command-to-column table lives in one Go file, and the client gets the emote by name on the event.
- The lines are this project's own wording, one pair (with and without a target) per emote.

**D7. Protocol, additive only.**
- `FlagJump = 1 << 1` on input runs.
- Requests: `posture` {`sit`: bool} and `sprint` {`on`: bool}; the slash commands use the same paths.
- Entity fields: `ps` ("sit", "air", "swim") and `vz`.
- Event: `emote` {from, to, x: name}.
- Chat channel: "emote".
- `effects` gains `spr` (sprinting).
- All of this goes in `docs/protocol.md` with a protocol test.

**D8. The Command window.**
- Built in `chrome.js` from the stock `command` template the way the menu bar is:
  - known events map to actions;
  - unknown events are greyed with a `MENU_WHY` style reason.
- It is shown at its default layout position and toggled with a key.
- Without `web/data/ui/` it is two plain buttons, Sit and Sprint, in the built-in HUD.

**D9. Keys.**
- Space jumps.
- X sits or stands, as in the stock client's sit toggle.
- Sprint gets a free key; R is already the walk toggle in `viewer.js`.
- Task 6.1 sets the final mapping after checking for clashes with the existing keys in `ui.js` and `viewer.js`. Shift stays the offline viewer's hold-to-sprint only.

## Risks / Trade-offs

- [A Step change breaks prediction for everyone] → New shared vectors land in the same commit as the Go and JS step, and `net.test.mjs` must pass before anything else uses the new fields.
- [Terrain noise reads as falls, and players take damage walking downhill] → `StepDown` threshold plus a census over each zone's start area, which fails if a walk along the navmesh triggers a fall over the safe height.
- [Water boxes are coarse, so a player floats over dry land at a box edge] → Swimming needs both "inside a box" and "surface above ground by `SwimDepth`", so shallow edges stay walkable.
- [Sitting regen factor wrong for 1.65] → Provisional status in the rule book. One constant to change.
- [Re-converting all player figures takes time] → Only the new roles are added: animconv merges into the existing `index.json`, as it did for cast clips.

## Migration Plan

- No store migration. Posture and sprint are not saved, so a player who logs out seated comes back standing.
- Re-run `animconv` over the player figures (the README recipe), then restart the server.
- An old client against the new server: unknown entity fields are ignored, and it can't jump or sit. A new client against an old server: the jump flag is ignored, and `/sit` is answered "Unknown command".
