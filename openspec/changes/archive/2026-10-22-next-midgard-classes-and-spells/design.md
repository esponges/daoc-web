## Context

`midgard-playable` ended with one playable class. What exists to build on:
- `internal/server/rules/character.go` has a `Classes` map with only `ClassWarrior`; `spec.go` and `abilities.go` carry the Warrior's lines.
- `internal/server/world/casting.go` and `rules/casting.go`, `rules/spells.go` already cast creature spells (10 of 15 creature spell types) with range, cast time, interrupts, recast and resists. Players do not use them.
- `catalog_spells.go` reads spells from the imported data; the census counts 73 class spell types, 0 implemented.
- The web client greys SPELLS with the tooltip in `web/chrome.js:16`.
- Characters are stored in `players.db`; every character is male.

## Goals / Non-Goals

**Goals:** a group of Warrior, Healer, Runemaster plus Berserker, Savage and Shaman in the Vale of Mularn; the playtest findings fixed or recorded with a source.

**Non-Goals:** the other classes; pets, charm, summon, realm abilities, advanced class choice at level 5; other realms; gender-specific armour art beyond what converters already produce.

## Decisions

1. **Reuse the creature casting engine, add a player front end.** The engine already resolves range, time, interrupt and resist. Player casting adds: a spellbook, a cast request, power spend and the client UI. Alternative: a second engine for players; rejected, it would drift from the creature one.
2. **Spell data comes from the import, rules from us.** Spell lines, levels, costs and effects come from the community database. The power formula and regeneration are ours, in `rules-power`, with sources recorded. A source for the power formula is searched for first (the rule sources list already used: disorder.dk, Camelot Herald wiki); if none gives numbers the rule is `provisional`, never copied from OpenDAoC-Core code (reading for understanding only, as CLAUDE.md allows).
3. **First slice of spell types, not all 73.** Types chosen by the five classes' counts in the census (direct damage, heal, damage over time, buffs, speed decrease, mesmerize, stun, bolt). Unsupported types stay in the book and say so. This keeps the census honest.
4. **Classes are data plus a small Go table.** Each class is one entry in `Classes` plus imported spec and style lists, following the Warrior. Class numbers and race lists come from the data and the ZAM start-zones source already adopted. Alternative: derive everything from the data; rejected, because the rule book wants the stats written in our words.
5. **Gender is a stored field with an additive protocol change.** `players.db` gets a `gender` column defaulting to male; `docs/protocol.md` gains the field and a protocol test. The viewer picks the female model set from `charconv`; if a race's female model is not converted yet the client falls back to male and the creation page says so.
6. **Loot table is ours and provisional.** The data has no generators (finding 1). We search once for a 1.65 source for drop rates by level; failing that, the table keeps the current 22% floor and adds level-banded items by creature kind, marked provisional. Not invented silently: the spec says provisional.
7. **Attack behaviour follows a source before it changes.** Findings 2 and 3 (toggle, auto-walk) are decided after reading how 1.65 behaved from the adopted sources; where none says, keep current behaviour and record it. The attack state on the bar is a UI fix either way.
8. **Night spawns: record, don't guess.** Staggering dusk spawns is only done if a source shows it.

## Risks / Trade-offs

- **Five classes at once is large.** Mitigation: tasks are tiered; the melee pair and the casting engine are independent and can ship separately.
- **Spell data quality.** Spells in the community database may carry post-1.65 types or values. Mitigation: mark them "not classic", as done for NPC kinds.
- **Power formula is provisional.** Healers may feel too strong or weak. Mitigation: provisional status, tuned in the integration playtest with the numbers recorded.
- **Female models may be missing for some races.** Mitigation: the fallback in decision 5, and a converter task to produce them.
- **`players.db` migration.** Additive only; a backup before the live restart, as in `midgard-playable` 4.1.

## Migration Plan

Back up `players.db`, re-run `dbimport` (staged store swapped at next start), rebuild and restart `gameserver`. Existing Warriors load unchanged as male characters with no power.

## Open Questions

- Which source gives 1.65's power pool and regeneration? (Two searches found nothing; the rule is provisional.)
- Does a source show the staggering of night spawns or the attack auto-walk? (Tasks 1.2, 1.3.)
- Are female models for every Midgard race converted by `charconv`? (Task 2.1 checks.)
