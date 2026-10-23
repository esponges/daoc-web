## Context

After `verify-spell-numbers` the pattern is set: a rule in `internal/server/rules` (our own words, a registry entry with sources and a status), a `rules-*` spec scenario per table test, wiring in `internal/server/world`, and a data check over the imported spell table. This change repeats it for the spell types that still do the table's number.

Today (observed in the code):
- A DoT tick is `SpellDamage(e.spell.Damage, 1, resist)` in `tickEffectsOf`, for players and creatures alike. The effect stores `caster` and `casterLevel` but no stat or spec.
- Life drain goes through the direct-damage case, so a player's drain already gets `PlayerSpellDamage` unless it is the no-variance type; the heal it returns is `DrainHeal(dmg, LifeDrain)`. Whether the sources scale drain this way is not yet checked.
- Damage add is `int(e.spell.Damage)` of the strongest effect (`Player.damageAdd`).
- Buffs are `BuffedStats(type, Value)` taking the strongest per stat, capped in `sheet.go` by `BuffStatCap` (level × 1.25) and `BuffAFCap` (level × 2, provisional).
- Debuff types are not in `ImplementedSpell`, so nothing here touches them.

## Goals / Non-Goals

**Goals:** a player's DoT, drain, damage add and buffs depend on the caster as the sources show; caps are sourced or explicitly provisional; the data check covers them.

**Non-Goals:** creature spells, debuffs, focus bonus on cost, the level-difference term (we keep DOL's, as decided before), protocol changes.

## Decisions

1. **Sources first, code second.** The first task reads the sources and records the formulas. If DOL and OpenDAoC disagree, follow DOL (as for damage) and record the difference. Nothing is copied: formulas are restated, and identifiers stay ours.
2. **Store the scale on the effect.** A DoT and a buff outlive the cast, so `addEffect` records the caster's scale (stat, spec, spell level) as `scaleFor` already computes it, and the tick uses it. Alternative: re-read the caster at each tick. Rejected: the caster may be dead or have changed gear, and a spell's strength is fixed when it lands.
3. **One scale function per behaviour, in `rules`.** For example a DoT per-tick amount function and a buff value function, each a pure function with table tests, rather than inline arithmetic in `effects.go`.
4. **Types the sources say do not scale stay flat.** If damage add or a buff is unscaled in both sources, the rule keeps the table's number, says so, and gets the status the source supports. That is a valid outcome, not a failure of the change.
5. **Statuses follow sources.** Reference where the community servers agree, provisional where nothing was found, documented only with a published 1.65 account. The earlier standard-search pass found only qualitative Grab Bag answers, so expect reference.
6. **Data check scope.** Extend `spellProblem` to the newly scaled types with the range the rule gives, as before. It still exercises the rules' functions, not the world's casting path; the world tests cover wiring.

## Risks / Trade-offs

- A community formula may be a later-patch behaviour (the combat regeneration note was one). Mitigation: record the patch hints found in the source comments and keep the rule at reference with the caveat.
- Scaling an effect at cast time makes a recast replace the old strength; the existing "restart, don't stack" rule already does that, and a weaker recast over a stronger one is an open question to settle from the sources (task 1.1).
- More rules at reference lower the share of documented rules in the census; that is honest, and the README table follows `parity.json`.

## Open Questions

- Do the sources scale a damage add by anything, or is it the table value? Settled by task 1.1.
- Is the armour-factor cap a function of level alone? Settled by task 1.1; if no source is found it stays provisional.

## Revisions after reading the sources (task 1.1)

- Only damage over time needs the caster's scale stored on the effect. Stat buffs and the armour buff are flat in the sources, and damage add depends on the swing and the caster's level, not their stat or spec, so decision 2 applies to DoTs alone.
- Damage add was a real bug rather than a missing scaling: the server added the table's `Damage` per blow as if it were a flat amount, where the sources treat it as damage per second times the swing, capped by level. The change covers it.
- The armour-factor cap stays provisional with an explicit entry, and the buff rule is split so the stat buffs can move to reference.
