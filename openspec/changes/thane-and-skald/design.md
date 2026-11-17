## Context

Six Midgard classes play today. The Thane and the Skald share the Viking base with the Warrior, Berserker and Savage, so their melee side is already there:
- their abilities (Shield, Protect, Guard, Intercept, Engage, Evade, the weapon kinds) are all marked implemented in `docs/parity`;
- trainers are matched by `Class.TaughtBy` through `importer.TrainerClass`, so the `ThaneTrainer` and `SkaldTrainer` kinds work once the classes are in `rules.Classes`;
- the data's start table (`startuplocation`, realm 2) gives Thane races 5, 6, 7 and Skald races 5, 6, 7, 8;
- `classxspecialization` gives the Thane `Stormcalling` and the Skald `Battlesongs`, both hybrid lines.

What the two spell lines ask for (read from the community dump):
- **Stormcalling (44 spells):**
  - 30 direct damage, 7 strength/constitution self-buffs and 7 self damage adds;
  - of the direct damage, 8 are Thunder Shout/Howl/Bellow/Roar (range 0, radius 300, instant, 20 s recast) and 4 are Mjollnir (range 1500, radius 350, 4 s cast);
  - the bolts and Toothgrinder spells are single-target.
- **Battlesongs (49 spells):**
  - 10 shouts (instant direct damage, range 700, 15 s recast), 5 mesmerize shouts and 5 speed decreases, all already supported;
  - 7 damage-add chants, 6 Songs of Rest, 5 Songs of Travel and 12 resist chants, all with `Target=Group`, `Pulse=1`, `Frequency=50`, `Duration=5` and no power cost.

The server today:
- `SpellDef` has no radius or pulse; `Radius` is already imported into `world.db`, `Pulse` and `PulsePower` are not;
- `castTarget` needs an enemy for an enemy spell, and the range check refuses a range-0 enemy spell;
- `speedFactor` only lowers speed, and `MoveInput.Slow = 1 − factor` is applied in `world/move.go` and `web/move.js`;
- `spellResist` is race plus items;
- health regeneration is `rules.HealthRegen` (provisional).

## Goals / Non-Goals

**Goals:**
- A Thane and a Skald can be made, trained and played to 50 with their whole lines working.
- Two new mechanics serve later classes too: area spells and pulsing spells.

**Non-Goals:**
- Hunter, Shadowblade, Bonedancer, Spiritmaster.
- Creature area spells.
- Debuffs.
- Instruments (Midgard has none).
- Realm abilities ("Call of a thousand storms").
- Hybrid-line "best spell per group" display: the book lists every spell, as for the other casters.

## Decisions

1. **The rules come from the sources first, as `verify-spell-numbers` did.** Before code, read DOLSharp and OpenDAoC-Core (understanding only, nothing copied) and search for a published 1.65 account for:
   - the class numbers (hit points, weapon skill base, spec multiplier);
   - area falloff;
   - how pulsing spells work (how many at once, what stops one, power per pulse);
   - speed enhancement's value, how it combines with slows, and whether combat breaks it;
   - the health regeneration buff's unit and whether it works in combat;
   - the resist buff's cap.

   Each goes in `findings.md` with project, file and search, and the rule's status follows. *Alternative:* take the community servers' behaviour wholesale. Rejected because the rule book wants each rule in our words with a status.
2. **Area targeting lives in player casting.** `continuePlayerCast` already fans a group spell out to targets. An area spell fans out to the creatures in the radius around the centre: the caster for range 0, else the target as the spell lands. A range-0 hostile spell skips `castTarget` and the range check. Each target goes through `castSpell` as now, so hit, resist, scaling and aggro are unchanged. *Alternative:* a separate area engine shared with creatures. Rejected for now; the decision on creature area spells is recorded for its own change.
3. **A pulsing spell is a song on the caster, not a long effect on the targets.** The player holds `song` (spell, next pulse). The zone tick re-casts it on the group targets each `Frequency` tenths of a second, and each landing is an ordinary effect lasting `Duration` and refreshed by `addEffect`. Leaving range or turning it off simply stops the refreshes, and the effect lapses. This reuses effect expiry and the effects message. How many songs a caster may keep on (one, or one song plus one chant) and what stops one come from the findings. Until those are read, the plan is one at a time, stopped by death, mesmerize, stun or turning it off.
4. **Speed above 1 goes through the existing `Slow` field.** A negative `Slow` (0.44 below zero for 1.44 times) already scales the step in Go (`speed * (1 - Slow)`). JS must do the same, and the held check stays `Slow >= 1`. A shared vector with a factor above 1 is added to the movement vectors both sides run. The wire already carries `speed` in the effects message, so no protocol field is needed for it.
5. **New effect kinds are additive values of `EffectView.Kind`:** `speed`, `regen`, `resist`. The book entry gains an additive `pulse` flag, and the effects message an additive `song` field (the spell id on, or absent). Both are documented in `docs/protocol.md` with a protocol test.
6. **Resists and regeneration read the strongest buff of each kind**, as stat buffs do:
   - `spellResist` adds the strongest resist buff for the damage type;
   - the health tick adds the strongest regeneration buff;
   - both are capped as the rules say.
7. **Class numbers are provisional entries in the class table** with a table test, following the five classes before them. Races and base class come from the data.
8. **"Call of a thousand storms" is marked not classic** in `cmd/parity`, with the reason, as creature kinds are, if the search confirms it is a later realm ability. Otherwise it stays counted as not implemented.

## Risks / Trade-offs

- **A pulse every 5 s over every group member is more work per tick.** Groups are at most eight, so the load is negligible at this scale.
- **The speed song changes movement prediction.** The shared vector and a browser check guard it. A mismatch would show as rubber-banding.
- **The community servers may disagree** on falloff or song limits, as they did on spell damage terms. The spec names the one followed and the findings record the disagreement.
- **An area Mjollnir pulling a whole camp** can kill a Thane quickly. That is faithful, not a bug.

## Migration Plan

1. Back up `players.db`.
2. Re-run `dbimport`. It adds the `pulse` and `pulse_power` columns, and the staged store is swapped in at start.
3. Rebuild and restart `gameserver`.

Existing characters are unchanged. A store without the new columns loads with no pulsing spells, so songs refuse to cast as unsupported rather than breaking.

## Open Questions

Each is settled in `findings.md` before code that depends on it:
- What are the hit points, weapon skill base and spec multiplier of the Thane and Skald?
- Does area damage fall off toward the edge in 1.65?
- How many songs or chants may a Skald keep on at once, and what stops one?
- Does a speed song break in combat, and how does it combine with a slow?
- Is the regeneration buff's `Value` hit points per tick, and does it work in combat?
- Is "Call of a thousand storms" a 1.65 ability?
