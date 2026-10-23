# Design

## Context

See proposal.md. Observed: `rules.SpellDamage(base, variance, resist)` multiplies a spell's `Damage` by a 0.75 to 1.25 variance and a resist factor; `Healed` adds the spell's `Value`; `BuffStatCap` and `BuffAFCap` are marked provisional. The registry marks `spell.damage`, `spell.buff`, `cast.heal`, `power.pool`, `power.regen` and `character.health-regen` provisional. `server/data/community/opendaoc-db-core` holds the spell table; the code repositories are not cloned.

## Goals / Non-Goals

- Goals: every spell number either traced to a named source or visibly provisional; mismatches against the data caught by a test.
- Non-Goals: new spell types, new classes, balance changes where a source is silent, copying any GPL code.

## Decisions

- **Clone the two repositories into a scratch directory** (the session scratchpad), read, and delete. They are never added to the repo or to `server/data`. Alternative: vendor excerpts as fixtures; rejected, it copies GPL code.
- **Sources are cited by project and file in `findings.md`, not in the shipping code's comments.** The spec carries the formula in our words and the registry's `Sources` field names `opendaoc` or `dolsharp`, as it already does.
- **Status moves only with a source.** A rule that a community server agrees with is `reference`; `documented` needs a published 1.65 account. Disagreement between the two servers is recorded and the DOL behaviour followed, as the closer to 1.65.
- **The data check is a table test over the real spell table**, skipping without the dump, like `TestParityDoesNotGoBackwards`. It reuses the importer's reader and `rules.SpellDamage`/`Healed` with fixed rolls, so it checks wiring, not balance.
- **Scaling comes after the answer.** If the sources show caster-stat or spec-level scaling, the amounts the check expects change with it; the data check (group 3) is written after the rules (group 2), against the confirmed formula, not today's.

## Risks / Trade-offs

- [The community servers have drifted from 1.65] → record disagreements, follow DOL, keep the status at reference.
- [Scaling is a large change] → measure it in findings first; split a separate change if it touches more than the rules package and the effects code.
- [A GPL formula is paraphrased too closely] → write the spec from the formula's shape and worked examples, not its code; no identifiers or comments carried over.

## Open Questions

- Whether the friends group's play (levels 1 to 30) is enough to defer high-level spec scaling.
