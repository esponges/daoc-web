# Design: spell-feedback

## Context

**Server events.** Spell events (`cast`, `spell`, `effect`, `resist`) carry the spell's name (`x`) but no id, and the client has no world store. Pet blows are already sent:
- `hit`/`miss` from the pet's NPC id (`npcHitsNPC`);
- blows on the pet from the creature's id.

The client drops both because `handleEvents` only logs lines where `from` or `to` is the player. The `pet` message carries the pet's id and its minions' ids (`PetMinion.ID`). The `effects` message carries `EffectView {nm, k, ic, left, of, cc}` for the player only.

**Client view.**
- `remote.js` plays `cast` on other entities, holding it until `spell`/`interrupt`/`castfail`. The player's own figure is driven by `viewer.js`, which never plays `cast`.
- No converted figure has the `cast` clip: `animconv` gained the "cast mid" role after the figures were converted.
- `particles.js` draws emitters whose state is a function of the clock and a slot hash. Emitter settings come from `internal/nif/particles.go` through `internal/scenery` (`propconv`).

**Install.**
- `gamedata.mpk` holds:
  - `spells.csv`: client effect id → name, icon, `Anim1–3`, `Effect1A`, `Effect1B`, `Effect2`, `Effect3`, projectile, target, buildup and pulse effect, and more;
  - `speffects.csv`: effect id → NIF id, attach node, duration (ms), rate, track and scale;
  - `spnifs.csv`: NIF id → file name.
- `effects/` holds the NIFs: 1,632 `.nif` files, many of them weapon glows we don't need.
- The cast animation ids come in triples (`32,33,34`, `35,36,37`, `47,48,49`…).

**Data.** The community `spell.ClientEffect` is the `spells.csv` row id; Minor Dusk Dart is 2501 in both.

## Goals / Non-Goals

**Goals:**
- Show every imported spell's caster, projectile and target effect as the client's tables describe them, and play cast animations on all casters.
- Log the pet's fight.
- Show the pet's effects and a readable effects row for the player.
- Fix the double cast.

**Non-Goals:**
- Sounds: the tables have a sound set column, but there is no audio system yet.
- Weapon glows and proc visuals on items.
- Area ground effects beyond the target effect at the area's centre.
- Spell icons by number: the sheets for most of the numbers aren't in this install.

## Decisions

### 1. The server sends the client effect, not the client looking it up

The event gains `fx` (int, omitted when 0), set from `spell.client_effect`.

*Alternative:* ship a spell-name → effect map to the client. That's rejected: names repeat across lines and levels, the server already knows the spell, and the change is additive.

### 2. `cmd/effectconv` converts only the effects the world uses

`effectconv` works like this:
- It reads the three tables from `gamedata.mpk` (with the existing `mpak` reader) and the set of client effects in `world.db`.
- It walks each effect to its NIF ids and converts each NIF once into `web/data/effects/<name>/`. Meshes use the same format `propconv` writes for models, and particle systems use `internal/scenery`'s emitter extraction.
- It writes `web/data/effects/index.json`:
  - `spells: { fx → { anims: [a1, a2, a3], caster: [effId…], projectile: effId, target: effId } }`
  - `effects: { effId → { nif, node, ms, rate, scale } }`
- It prints a census of node types it couldn't convert. This is the spec's "cannot draw" scenario.

Converting only the effects the world uses keeps the output to the few hundred effects Midgard's spells reference, not 1,632 files.

### 3. One-shot effects on the existing particle path

`particles.js` is clock-driven per placement. An effect instance is a placement with `start = now` that is dropped after `ms`. The hash seed is per instance, so two casts don't fire in step. Meshes in effect NIFs draw with the props shader, additive when the NIF's alpha property says so. Instances attach to a bone by name through the figure's skeleton (`Bip01 L Held`, `rootnode` = feet). The bone's matrix is read each frame, so effects follow movement.

A projectile moves from the caster's hand bone to the target's chest along a straight line over the time between the `spell` event's arrival and the hit. Its speed is fixed, and its trail is the projectile effect's emitter.

### 4. Cast animation ids

`spells.csv`'s `Anim1–3` name the caster's start, loop and release. Task 2.1 first reads the client's animation tables (`anims.csv`/`canims.csv`, as `animconv` uses) to find what these ids index. Then `animconv` converts `cast_<id>` clips for every figure, keeping the generic `cast` ("cast mid") as the fallback.

The viewer plays them for the player's own figure on `cast` (start, then the held loop) and `spell` (release), and clears on `interrupt`/`castfail`. `remote.js` does the same for others with the `fx` → anims lookup.

If the ids turn out not to resolve, the generic clip is played. That keeps the spec's fallback and changes no other part of this design.

### 5. The pet log is client-side

`handleEvents` treats `from`/`to` equal to `net.pet.id`, or to any `net.pet.mn[].id`, as "your <pet name>". It uses the same line builders with the pet as subject or object (`effects.js` `petLine`). There's no server change: the events already reach the owner, because they're sent in range and the pet is near its owner.

### 6. Pet effects on the pet message

`protocol.Pet` gains `Fx []EffectView` (`fx`, omitted when empty), filled from the pet NPC's effects list in `petMsg`. `petMsg` is already re-sent when the pet changes. It also has to be re-sent when an effect starts or ends on the pet, which is a call in the effect add/remove paths for pets. The pet window draws the six with the least time left.

### 7. Own effects row

This replaces the `#effects` badges:
- 26-px icon frames drawn by kind, using the existing `EFFECT_KINDS` colours inside a stock frame piece when the stock UI is present;
- a harmful frame (red) or helpful frame (gold) by kind;
- time under each icon.

It's placed under the summary window, computed from the summary's rectangle rather than a fixed `top: 150px` (today that sits under the clock).

### 8. One cast per press

`castSpell(id)` is ignored while `net.me.casting` is set or within 300 ms of the last request for the same spell. A key repeat (`e.repeat`) is ignored on quickbar keys.

## Risks / Trade-offs

- **Unsupported controllers:** effect NIFs may use controllers our NIF reader doesn't support (texture flipbooks, alpha and colour keys). → The census reports them. A static fallback (first frame) is drawn instead of nothing, and the biggest gaps are listed in findings for a follow-up.
- **Bone names:** a figure may not have the bone the table names. → Fall back to the figure's root at its chest height.
- **Performance:** many casters in one place add up. → One draw per emitter per instance, a cap of 64 live instances with the oldest dropped first, and no effects past the drawing distance.
- **Re-import and re-conversion:** players must run `dbimport`, `effectconv` and `animconv` again. → The README run-order gains the two converters, and the client works without them (spec fallback).

## Migration Plan

1. Deploy the server with the column migration. Old world stores load with `fx` = 0 (spec scenario).
2. Re-import, then run `effectconv` and `animconv` over all figures.
3. Rollback is a revert. The extra column and the extra JSON fields are ignored by older code.

## Open Questions

- What exactly `Anim1–3` index in the client's animation tables. This is resolved in task 2.1; the fallback keeps the design unchanged.
