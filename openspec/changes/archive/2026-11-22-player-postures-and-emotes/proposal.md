# Proposal: player-postures-and-emotes

## Why

A player character today can only walk, run, strafe, fight, cast and die. Everything else the 1.65 body does is missing:

- **No sitting.** In 1.65 a player sits to rest: the stock Command window has a SIT button (`ToggleSitting`), and resting while seated is how downtime worked before patch 1.87. Here there is no sit command, no seated state on the server and no seated animation. The regen rules already have a "Resting" scenario, but nothing distinguishes sitting from standing.
- **No jump, no fall, no falling damage.** `Step` (`world/move.go`, `web/move.js`) snaps Z to the ground every step, so a character glides down cliffs. Patch 1.38 turned falling damage on and added Safe Fall to the Stealth line, and both are part of 1.65.
- **No swimming.** Players walk along riverbeds and lake bottoms. The client has swim clips ("Crawl", `Sw_h.kfa`) and the zone data knows where the water is (`static.go IsWet`, `man.waters`).
- **Sprint exists only offline.** The offline viewer has Shift to sprint at 420 units/s. The server knows nothing of it: no endurance drain, no speed. The stock Command window has a SPRINT button.
- **No emotes.** There's no `/wave`, `/bow` or `/dance`. The stock chat window has an emote button and an "Emote" channel, and the client's command table lists them.

The install already carries what this needs. For every player anim set, `anims.csv` names:
- Jump and Land: `ju_h_a`, `ju_h_b`;
- Sit Down, Sit Idle and Stand Up: per race, e.g. troll `s_tm_a/b/c`;
- the swim clip;
- all the emote clips: `em_wavelong`, `em_bow`, `em_rvrdance`, and so on.

`animconv` simply never converts them.

## What Changes

- **Sitting (server-owned posture):**
  - `/sit` and the stock SIT button toggle sitting; so does the sit key.
  - The server keeps the posture. Moving, attacking, casting, shooting, or being hit or attacked stands the player up, with the game's messages.
  - A seated player regenerates faster: the rate is a rule-book entry (`rules-resting`), provisional until a 1.65 source gives the number.
  - Other clients see the sit-down, seated idle and stand-up clips.
- **Jumping and falling:**
  - The movement step gains vertical motion: a jump impulse, gravity, and landing on the ground.
  - It's the same maths in Go and JS, with new shared vectors.
  - The jump key sends a new input flag (additive).
  - A long drop costs hit points under a new `rules-falling` spec. Safe Fall (Stealth line) raises the safe height and lowers the damage.
  - Jump and Land clips play.
- **Swimming:**
  - Under a water surface the body floats at the surface instead of walking the bottom, moves at the swim speed (`rules-movement`), and plays the swim clip.
  - Players can't sit or fight in some ways while swimming; the spec says which.
- **Sprint:**
  - Sprint is a server-side toggle from the stock SPRINT button, `/sprint`, or the key.
  - It raises run speed and drains endurance each second under a `rules-movement` entry, and stops when endurance runs out.
  - Prediction uses the speed factor the server sends, as slows already do.
- **Emotes:**
  - `/<emote>` (with or without a target) plays the emote clip on every nearby client and sends the emote line ("You wave." / "Thora waves at you.") on the Emote chat channel.
  - The set is the emotes the 1.65 client had. Later emotes the table also lists are left out, not guessed (`design.md` D6).
  - Emotes can't be used in combat or while dead.
- **Animation conversion:** `animconv` takes the new roles from `anims.csv` (jump, land, sit down/idle/up, swim, and each emote clip). Emote clips are written lazy, like cast clips. All player figures are converted again.
- **Stock Command window:** built from its template. SIT and SPRINT work, ATTACK, PRAY and RELEASE map to existing actions, and FOLLOW, STICK, FACE and BONUSES are greyed with a reason. A plain fallback is used when `web/data/ui/` is absent.
- **Protocol (additive):**
  - input flag `FlagJump`;
  - request `posture` (`sit`/`stand`) and `sprint`;
  - entity fields `ps` (posture: `sit`, `swim`, `air`) and `vz` (vertical speed, for the jump clip);
  - event `emote` with the emote id and target;
  - the `effects` speed factor covers sprint.

## Capabilities

### New Capabilities
- `player-postures`: sitting and standing, jumping and falling, swimming and sprinting as server-owned states of a player's body. Covers their animations on every client and the controls that change them.
- `emotes`: the classic emote set: commands, the animation every nearby client sees, the emote chat lines, and when emotes are refused.
- `rules-resting`: regeneration while seated, compared to standing out of combat.
- `rules-falling`: when a fall hurts, how much, and Safe Fall.
- `rules-movement`: sprint speed and its endurance cost, swim speed, and the jump and gravity numbers the movement step shares.

### Modified Capabilities
- `world-sync`: the movement step gains vertical motion (jump, gravity, landing, swimming at the surface). It stays one definition shared by Go and JS with shared vectors, and the input gains the jump flag.
- `rules-character-stats`: health and endurance regeneration take the seated bonus from `rules-resting`.
- `rules-power`: power regeneration takes the seated bonus.
- `chat-channels`: emote lines arrive on the Emote channel and are coloured and tabbed as the stock chat window does.
- `stock-ui-windows`: the stock Command window is built and its buttons are wired or greyed.

## Impact

- **Rules:**
  - new `internal/server/rules/resting.go`, `falling.go` and `movement.go`, each with table tests and registry entries;
  - specs `rules-resting`, `rules-falling` and `rules-movement`;
  - new sources go in CLAUDE.md and the README: 1.38 patch notes for falling and Safe Fall, and whichever source pins the sitting bonus.
- **Server:**
  - `world/move.go`: Body gains VZ and an air/swim state, Keys gains Jump, `Step` gains gravity and water;
  - `world/zone.go`: regen reads the posture;
  - a new `world/posture.go` for sit, stand, sprint and the stand-up triggers in combat, casting and archery;
  - `world/emotes.go`;
  - `world.go` adds the commands;
  - fall damage on landing.
- **Protocol:** `protocol.go`, `docs/protocol.md`, a protocol test, and regenerated `testdata/move_vectors.json`.
- **Client:**
  - `web/move.js` mirrors `Step`;
  - `net.js` sends the jump flag and predicts sprint speed;
  - `viewer.js` and `remote.js` choose the posture clips and play emotes;
  - `ui.js` adds keys and the command window, and `chrome.js` the Command window;
  - `effects.js` adds emote and fall log lines.
- **Converters:** `cmd/animconv` gets the new table roles and a re-run over every player figure. Creatures are not touched.
- **Data:** nothing committed. Converted clips go to `web/data/char/*/anim/` (gitignored).
