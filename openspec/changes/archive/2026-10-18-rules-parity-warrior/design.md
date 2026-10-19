# Design

## Context

Today's rules live in `internal/server/world/rules.go`, as functions the zone calls with ad-hoc arguments: `playerMaxHP(level)`, `weaponDamage(dps, speed, level)`, `npcDamage(level)` and the rest. They draw randomness from the zone's `rng` in the middle of `combat.go`. Players have no stats, specs or abilities. A character row has only race, class, level, experience, hit points, endurance and money (`internal/server/store/store.go`). Creatures get hit points and damage from their level alone, and their template's parry, evade and block (`npc.go`).

The world store already holds more than the server reads:
- `spawn.classtype`, `spawn.brain`, `item.spd_abs`, `item.quality`, `item.level_req`, `style.armor_loc`;
- `class_spec`, imported from `classxspecialization`, which gives the Warrior's specs: Axe, Hammer, Sword, Parry, Shields, Thrown Weapons, plus "career" rows with negative levels.

It does not hold:
- item stat and skill bonuses (`Bonus1..10`, `Bonus1Type..10Type`);
- the item enchantment `Bonus`, condition and durability;
- the spawn's race;
- the race resists;
- `specxability`, which says which abilities the careers and specs grant at which level, for example MidArmor 4 (chain) at Warrior career level 10, or Evade 1 at 15.

The formulas in the specs were worked out from public mechanics pages, with the classic reference server (OpenDAoC-Core) read to understand what no public page states. The code that implements them is written from the specs.

## Goals / Non-Goals

**Goals:**
- One place holds every formula, testable without a zone: inputs in, numbers out, rolls supplied.
- The Warrior plays by 1.65 numbers end to end: creation, training, fighting, levelling, dying.
- The census can say, from the data, how much of Midgard the server covers, and the number can only go up.

**Non-Goals:**
- Changing how NPCs pick targets, roam, respawn or call for help (the next change, "NPC as written").
- Group mechanics (the attacker count used for evade and parry counts creatures only).
- Stock trainer and healer windows. Built-in panels come first; the converted stock windows follow in a UI change and fall back to these panels.
- Performance work. The rules are plain arithmetic per swing.

## Decisions

### 1. A pure `internal/server/rules` package

The package holds:
- races and classes as Go tables;
- stat sheets;
- melee resolution;
- creature numbers;
- experience and con;
- death penalties.

It imports nothing from `world`. Randomness enters as a `Rolls` value: a source of uniform floats the caller provides. `world` passes one backed by the zone's `rng`, and tests pass fixed sequences. The main entry is a swing resolver. It takes an attacker and a defender description and returns the outcome:
- the kind of result;
- the location;
- the damage;
- whether it was critical;
- whether the style fired;
- the endurance spent.

The descriptions cover level, stats, weapon, spec, defences, armour by location and resists. Players and creatures fill the same description, so a creature hitting a player and a player hitting a creature use one code path.

*Alternative:* keep the functions in `world` and add parameters. That was rejected, because the formulas need a dozen inputs each, and the rule-book vectors need to run without a zone, a catalog or a store.

### 2. A rule registry the census and tests share

Each rule is a registry entry in the rules package:
- an id such as `melee.parry`;
- the capability and requirement it implements;
- its status (documented, reference or provisional);
- its sources.

The census reads the registry to decide what is implemented and how many rules of each status exist. The vector tests are table tests named after the scenarios, for example `TestRules/rules-melee/Parry 20`. A test that walks the change's and the main specs' `rules-*` files and checks that every numbered scenario has a test is a good check, but brittle. Instead, the registry lists its scenario names, and a test asserts that each named test exists. That gives the same guarantee without parsing Markdown.

### 3. Class and race definitions are code; abilities are data

The numbers that OpenDAoC keeps in C# classes are written as Go tables, in our words:
- base hit points;
- weapon skill base;
- the spec multiplier;
- the stat order;
- the eligible races;
- the race base stats.

The abilities granted by level come from `specxability` and `classxspecialization`, The importer copies `specxability` as it is into a new `spec_ability` table: the source (a career or a spec), the level needed, the ability key, the ability level and the class (0 for any). The server joins it with `class_spec` (already imported) to find each class's grants. A key ending in "Career" is a career, granted by character level; any other key is a spec, granted by its trained level. The server asks which abilities a character has at their level and specs, and implements only these:
- the weapon families;
- the armour level;
- the shield size;
- Evade.

Other abilities are carried as known but inactive, for the sheet and the census. Item object types map to ability needs in one table. For Midgard:
- 11 sword, 12 hammer, 13 axe and 8 staff need the matching Weaponry ability;
- 32 to 35 (cloth to chain) need armour level 1 to 4;
- 42 is a shield, whose size is in its damage-type field (1, 2 or 3) and must not exceed the Shield ability.

The importer's `Hand` is 1 for two-handed items. Today's code treats 2 as two-handed, which is the left hand. That is fixed here.

### 4. The stat sheet is derived, not stored

What a character is, and what is saved:
- base stats chosen at creation;
- spec levels;
- level and experience;
- constitution lost;
- death count;
- the gravestone.

Everything else is computed into a `Sheet` whenever level, specs, gear or constitution change:
- modified stats;
- maximum hit points;
- the item caps applied;
- weapon skill and armour factor for display;
- unspent points (earned by level minus spent on specs);
- abilities.

`Player.maxHP()` and the swing code read the sheet. This replaces `recalcGear`'s weapon, shield and armour fields with a fuller recompute, done at the same moments.

*Alternative:* store unspent points. That was rejected because two numbers that must agree can disagree. Earned points are a pure function of level and half-level.

### 5. Store migration

The migration is additive, like the existing `gender` column, and runs in `store.Open`:
- `character` gains `str, con, dex, qui, int, pie, emp, cha` (NULL for old rows), `con_lost`, `death_count`, `xp_scheme` and gravestone columns: `grave_zone`, `grave_x`, `grave_y`, `grave_z`, `grave_xp`.
- A `char_spec(char_id, spec, level)` table holds trained specs; a missing row means level 1.

`xp_scheme` is 0 for rows written by the old curve and 1 for the new table. On first load of a scheme-0 character, the world converts their experience to the same fraction of their level under the new table. It keeps the old curve's formula in the migration code only, saves, and sets the scheme to 1. NULL base stats are filled with the race base plus the default build (+10 to strength, constitution and dexterity) on first load.

Creation accepts an optional build. Without one it applies the default, so `login.js` keeps working unchanged.

### 6. World store additions

The world store schema gains:
- `item`: `bonus` (the enchantment percent), `condition`, `max_condition`, `durability`, and a child table `item_bonus(item, type, amount)` for `Bonus1..10` with their types. Only stats (1–8), hit points (10), the Midgard weapon and defence skills (40 parry, 43 shields, 52 sword, 53 hammer, 54 axe, 55 left axe, 91 thrown) and the all-melee-skills bonus (164) are read today. The rest are kept, for the census to count.
- `spawn`: `race`.
- `race`: the resist columns.
- `spec_ability`, as above.

`dbimport` must be re-run. The server detects an old store from missing columns, logs that it should be re-imported, and falls back to no bonuses, no race and no abilities. It does not fail, matching how class types were introduced.

### 7. Combat integration

`combat.go` keeps everything around the swing: timers, arcs, targets, events, experience and loot. The swing itself becomes one call to the rules package. Some fixes go in with it:
- The "seen" test for defences uses exactly the 120° cone the specs name, instead of `frontArc × 1.3`.
- The attacker count for a player is the number of creatures whose target is them and that are within melee reach.
- A style's endurance is charged when the swing resolves: full cost if it fires, half if the swing failed. It is not charged when it is queued; queueing only checks that the full cost is in hand.
- Swing time comes from the rules (quickness), for players and creatures.
- Creatures are built from `rules.Creature(level, template, race, two-handed?)` at spawn. `npcSwing(size)` and the invented curves go.

Experience moves to the rules package: shares, cap and grey check. The loot owner stays the top damager. `killMoney` stays as is: money drops are not part of this change, and it becomes a provisional rule in the registry so the census shows it.

### 8. Death, gravestones and healers

At death, the first-stage experience loss applies at once. Release, by request or by the existing timer, applies:
- the second-stage loss;
- the gravestone;
- the constitution loss;
- the return to the bind point with full health.

A gravestone is a world object owned by the zone it is in, keyed by its owner. It is sent as an entity of a new kind `k:"g"` with the owner's name. It is persisted on the character, and recreated on server start for characters whose grave is in a loaded zone.

Praying is a five-second timed action, cancelled by movement input, an attack or being hit. The client asks with `pray`, which is also typed as `/pray` with the grave targeted.

Healers answer `interact` with `healer`, which carries the points lost and the cost. `restore` pays for as many points as the money covers, in one store transaction with the character save.

### 9. Protocol (additive, version stays 1)

| message | direction | what it carries |
| --- | --- | --- |
| `stats` | server to client | the sheet: base and modified stats, max HP, max endurance, specs and their levels, unspent points, abilities (with an `on` flag for implemented ones), weapon skill and AF as displayed, constitution lost |
| `trainer` | server to client | the trainer NPC's id and name, plus the spec rows: name, level, cost of the next level, and whether it can be trained now |
| `train` | client to server | `spec` and `n` (levels to add, default 1); the reply is `trainer` again plus `stats` and `bar` |
| `healer` | server to client | NPC id, points lost, price per point, total |
| `restore` | client to server | restore lost constitution at that healer; the reply is `inv`, `stats`, and a chat line |
| `pray` | client to server | pray at the targeted gravestone |
| entity `cc` | in `snap` | the con colour −3..3 for creatures, per receiving player |
| entity `k:"g"` | in `snap` | a gravestone; `nm` is "grave of <name>", `own` is true for the viewer's own |
| events | in `ev` | `xploss` (amount), `conloss` (amount), `pray` (started, done or cancelled), `train` (spec, new level) |

`ik` gains `trainer` and `healer`. Each message gets a protocol test and a line in `docs/protocol.md`.

### 10. Client

Built-in panels follow the plain-panel style of the existing fallbacks:
- **Character sheet:** opened with C or from the developer panel; shows stats, hit points, weapon skill, AF, specs and abilities.
- **Trainer:** spec rows with a Train button, plus the unspent points.
- **Healer:** a dialog with the price and Accept.

The gravestone is drawn as a small upright marker. A converted grave model can replace it later. The target window colours creature names by `cc`, using the game's con colours. Death messages for experience and constitution loss go to the chat log.

### 11. The parity census

`cmd/parity` reads the dump with `internal/importer.ReadTable`, so no new parser is needed. It reads:
- `mob` and `npctemplate`;
- `spell`, `spellline` and `linexspell`;
- `classxspecialization`, `specxability` and `style`;
- `stylexspell`;
- `classxrealmability_atlas`;
- `dataquest`, `merchantitem`, `door`, `keep`, `relic`, `battleground`.

Midgard means:
- spawns in regions 100, 101 and 151;
- the 1.65 Midgard classes and their base classes (Viking, Mystic, Seer, Rogue).

It writes, sorted and without timestamps:
- spell types with use counts (by creatures, and by each class);
- style features: openings, positionals, chains, procs;
- abilities by class;
- NPC class types with placements;
- services;
- rule statuses.

It is stored in `docs/parity/parity.json` and `docs/parity/README.md`. The regression test runs the census when the dump is present (skipped with `-short`, like the asset census) and compares implemented flags and statuses with the committed baseline. Counts of data may change with a new dump without failing; an implemented item becoming unimplemented, or a documented rule becoming reference or provisional, fails.

### 12. Sources

These are recorded in CLAUDE.md and the README under "Where things come from", and named in each rule:
- disorder.dk's 1.65 mechanics pages (chance to hit, evade, parry, blocking, stats and caps, hit point cap, armour factor cap, swing speed, hit location, skill cap);
- the Camelot Herald wiki (melee damage, styles);
- Mythic's Grab Bag answers and the knowledge-base note on the 1.25 experience clamp;
- the published experience and constitution-restore tables, via silicondragon.com;
- the Mythic developer post on con colour bands;
- OpenDAoC-Core, read to understand (reference status), never copied.

## Risks / Trade-offs

- [The game gets harder at once: level 1 hit points fall from 175 to 39, styles need training, and death costs experience] → Expected: this is the point. The README's combat section says so. The rule statuses show which numbers are ours to tune. Creatures near the Myrkwood spawn will be more dangerous, which `stock-ui-followups` already notes.
- [Reference-status formulas may not match live 1.65] → They are marked as such. The census shows how many rules depend on the reference. A later change can add the local reference-server harness to compare outcomes over many swings.
- [The con band rule differs from the reference server below level 40] → Recorded as provisional. It changes which low-level kills are grey, which is a small amount of experience either way.
- [Hit location weights differ between sources (40/25/15/10/5/5 against a measured 44/22/16/12/3/3)] → The reference weights are taken, and the measurement is recorded in the rule.
- [An old `world.db` lacks bonuses, races and abilities] → The server warns and runs without them: no abilities means a Warrior can only wear what needs no ability, which the log makes obvious. The README says to re-run `dbimport`.
- [Experience migration loses precision] → Fractions are computed in float64 and rounded down. Characters are few and private.
- [The parity census reads the GPL dump] → It writes only type names and counts, the same kind of output as the asset census.

## Migration Plan

1. Re-run `go run ./cmd/dbimport` to rebuild `world.db` with bonuses, races, resists and `spec_ability`.
2. Restart `gameserver`. `players.db` gains its columns on open, and each character converts on first join.
3. Run `go run ./cmd/parity` and commit `docs/parity/` with the change.

Rollback: an older server build ignores the new columns and table and still opens `players.db`. A character's experience after conversion is on the new scale, so playing them on an old build would put them far above their level on the old curve. Back up `players.db` before the first run.

## Open Questions

- The provisional regeneration rates (health every 3 s out of combat, endurance 6 and 2 a second) can be tuned without changing the specs' structure, once a better source is found.
- Whether the gravestone uses a model from the install can be decided when the marker is drawn. The spec only needs a visible, targetable object.
