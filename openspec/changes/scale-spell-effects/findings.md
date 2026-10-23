# Findings

Read 2026-10-23 from DOLSharp (`Dawn-of-Light/DOLSharp`) and OpenDAoC-Core, shallow clones in the session scratchpad, not in this repo. Formulas are restated in our own words from their shape; no code or identifiers are carried over. File names are for finding them again (all under `GameServer/`).

## 1.1 What the community servers do

### Damage over time (`spell.damage-over-time`) — `spells/DoTSpellHandler.cs`, `spells/SpellHandler.cs`

- **Base damage** is the spell's `Damage` times `(casting stat + 200) / 275` for a player with a mana stat, the same factor as direct damage (DOL's DoT handler restates the base calculation; poisons, item effects and a few classes are exempt). Not scaled by the casting stat for creatures here (ours stay as they are).
- **Variance is rolled on every tick**, not once. DOL: a range from `0.13 + (spec − 1) / target level` up to `1.13`, the low end clamped between 0 and the high end. There is **no level-difference term** for a DoT in either project. Cap: three times the spell's `Damage`, the same cap as direct damage. Resists apply before the cap.
- **Disagreement.** OpenDAoC drops DOL's DoT-specific range and uses the general spell variance (the one our direct-damage rule follows) with the level-difference term zeroed for DoTs. We follow DOL for the range, since it is the older design and states the DoT numbers outright, and note the gap here.
- **This server today** hits for the table's `Damage` on every tick with no variance, for players and creatures alike.
- **Recast.** Both projects replace a running DoT of the same kind when the new one lands. Our existing rule (the same spell restarts, not stacks) stays; with scaling, the restart also takes the new cast's strength.

### Life drain (`spell.lifedrain`) — `spells/LifedrainSpellHandler.cs`

- A drain is built on the direct-damage handler, so a player's drain scales exactly like direct damage; the caster is healed the damage done times the spell's return percent (halved while diseased). **This server already routes a player's drain through the player damage rule** (done in `verify-spell-numbers`), so nothing new is needed; the findings record it so the data check can cover it.

### Damage add — `spells/DamageAddAndShield.cs` (DOL), the same file in OpenDAoC

- **The spell's `Damage` is a damage per second, not a per-blow amount.** DOL: on each landed blow the add is `min(Damage, cap) × swing seconds × (0.5 to 0.99)`, the roll uniform in hundredths (a base of 50 plus up to 49). The cap on the per-second figure is `(1.2 + 0.3 × caster level) × 0.7`: about 1.47 at level 3, 5.04 at level 20 and 7.14 at level 30. A **negative** `Damage` is that percent of the blow's own damage instead.
- OpenDAoC: per-blow add is `Damage × swing seconds` with the general spell variance (0.75 to 1.25 by default) and a clamp to the most a swing could have done; it states "no documented source" for the percent form.
- **No casting stat and no spec scaling in either.** The caster's level matters only through the cap.
- **Disagreement.** The roll (DOL 0.5–0.99, OpenDAoC about 0.75–1.25). We follow DOL, where the cap and spread are stated, and record the gap.
- **This server today** adds `int(Damage)` flat on every landed blow, ignoring swing speed and the cap. For a 3.0 second weapon and a table value of 1.6 the add at level 3 (cap 1.47) should be about 2.2 to 4.4, not 1. **A real bug** (the table's own description says "additional damage" without the per-second meaning).

### Buffs — `spells/PropertyChangingSpell.cs`, `spells/SingleStatBuff.cs`, `propertycalc/StatCalculator.cs`

- **Stat buffs (strength, constitution, dexterity, quickness and the dual ones) apply the spell's `Value` times an effectiveness of 1** in DOL: no caster stat and no spec scaling. A player's total from baseline buffs is capped at `level × 1.25`; specialisation-line buffs have a separate cap of `level × 1.5 × 1.25`. Both projects agree. Our `BuffStatCap` already matches; no new scaling.
- **Disagreement (later behaviour).** OpenDAoC scales armour-factor buffs and all debuffs by the caster's spec in the line: effectiveness `0.75 + (spec − 1) × 0.5 / spell level`, clamped to 0.75–1.25 (list casters excluded). DOL does not. We follow DOL (flat) and note that this is a modern OpenDAoC choice with no 1.65 source found.

### Armour-factor buff cap (`BuffAFCap`) — `propertycalc/ArmorFactorCalculator.cs`, `gameobjects/GameLiving.cs`, OpenDAoC `gameobjects/GamePlayer.cs`

- The two projects treat the armour-factor buff differently and neither states a flat "buffs add at most N" cap. DOL caps the specialisation-line armour buff at `level × 1.875` and leaves others uncapped; OpenDAoC applies the baseline armour buff piece by piece (a fifth of the buff to each worn piece, with each piece's own armour-factor cap applied again). The absorb code in DOL carries a separate, absorb-only cap.
- **Conclusion: no single sourced cap.** `BuffAFCap` (level × 2) stays provisional. It matches a level × 2 item cap per piece in spirit, not as a buff cap. The registry keeps a provisional entry for it and `findings.md` records why.

## 1.2 Published 1.65 accounts (2026-10-23, standard searches)

- **Damage over time.** Query: Dark Age of Camelot damage over time spell damage spec level casting stat variance 1.65. Result: nothing specific to DoTs. Mythic's Grab Bag answer (already used for direct damage) says spec above a spell's level gives full effect; a Camelot Herald page for the Sorcerer says DoTs can be stacked, with no formula. No figures. DoT scaling stays reference (the community servers).
- **Damage add and buff caps.** Query: damage add proc damage per second weapon speed cap level buff cap 1.25 level armor factor buff. Result: nothing on live 1.65. One private server (Storm) lists its own caps ("Bonus to AF Cap is Level * 2"), which is not live evidence and is not used. A tool page notes quickness does not apply to damage add. No source for the damage-add formula or the AF buff cap. Damage add stays reference (the community servers); the AF cap stays provisional.
- **Casting-stat factor and spec scaling of buffs.** Not found (see the earlier `verify-spell-numbers` findings for the stat factor).

## 1.3 Open questions settled

- **Does damage add scale?** Not by caster stat or spec, in either project. It scales with the weapon's swing time and is capped by the caster's level (rule above).
- **Does a weaker recast replace a stronger one?** For a DoT of the same spell, yes: the recast restarts the duration and takes the new strength (both projects replace). Different DoT spells on one target stay as the existing rule has them (not changed here).
- **Is the armour cap level-only?** The sources do not state a flat cap; see above. Stays provisional.

## What the game showed (2026-10-23)

Seen on a real server on a scratch port with a scratch data copy, driven over its HTTP and WebSocket interface from a Node script (the Chrome extension was not connected this time, so the rendered UI was not looked at).

- **Damage over time.** A Norseman Shaman (level 40, casting stat 94) cast Fungal Solvent (damage 19, spell level 11) on a level 11 corpse eater. At a Subterranean spec of 11 the six ticks were 22, 23, 21, 22, 22, 22 (the rule's range for a level 11 target is 1.04 to 1.13). At a spec of 20 every tick was 23, the top of the range, as the rule gives (the low end is clamped to the high end). The old code did exactly 19 on every tick.
- **Stat buff.** Lesser Strength of the Viking (value 11) raised the Shaman's strength from 70 to 81 at a spec of 20, the spell's value, whatever the caster. Not observed at a low spec in the game; the flat value is covered by the rules test.
- **Damage add** was not cast in the game (it needs a melee target and a blow); the world test `TestOverTimeSpellsScale` covers it.
- **A bug found on the way, not fixed here: buffs from the data end at once.** The stat and armour buffs of the Shaman and Healer lines (for example spell 3161) have a `Duration` of 0 in the table: they are concentration spells, which last until cancelled and draw on the caster's concentration, and the server gives a zero duration an end time of now, so the buff vanished a few seconds after the stat rose. Concentration is on the README's "not yet" list. It wants its own change (`concentration-buffs`): a buff with no duration lasts until cancelled, death or logout, with the concentration limit and a way to cancel it.
