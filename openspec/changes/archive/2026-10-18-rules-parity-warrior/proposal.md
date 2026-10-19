# Proposal

## Why

The goal is to play Midgard as it was on a classic 1.65 public server: same rules, same creatures, other people in the world. Today the world and its population are imported, but every number that decides a fight is made up. `rules.go` gives a Warrior `120 + 55 × level` hit points, scales damage by an invented factor, and gives creatures hit points and blows from a curve chosen to "feel like a fight". Nothing measures how far the server is from the real game. This change starts closing that gap with the Warrior, the class players already have, and puts in place two things every later step needs: a rule book that keeps the rules our own, and a census that turns "distance from a public server" into a number.

This is the first change of the Midgard parity plan, which runs: rules core and a real Warrior (this change), then NPC behaviour as written in the data (respawn, aggro, factions), creature magic, town services, character creation with all Midgard classes, player magic (Thane, then Healer), groups and guilds, and later the other realms and RvR.

## What Changes

- **Rule book.** Each game mechanic the server implements gets a capability spec under `openspec/specs/rules-*/`, written in this project's words, with the maths, the sources it came from, and scenarios that double as test vectors. A rule is marked *documented* when a public source states it, *reference* when it matches how the classic reference server (OpenDAoC, read only to understand) behaves and no public source gives the number, or *provisional* when neither does and the value is our choice. CLAUDE.md and the README list the sources adopted. No code is copied from a GPL server.
- **Parity census.** A new command, `cmd/parity`, reads the community dump and the server's own registry of implemented rules. For Midgard it reports what the data asks for (spell types and how many spells use each, style features, abilities, creature behaviours, NPC services) against what the server implements, plus how many rules are documented, reference or provisional. It writes `docs/parity/parity.json` and `README.md`, which are counts and names only, and a test fails if a count goes backwards.
- **A rules package.** The formulas move out of `internal/server/world/rules.go` into a new package, `internal/server/rules`. It is pure and deterministic: no world, no clock and no randomness of its own (the caller passes the rolls), so every scenario in the rule book is a table test.
- **A real Warrior.** These replace the invented numbers:
  - **Stats:** race base stats (Norseman, Troll, Dwarf, Kobold, Valkyn), 30 creation points, level gains from 6 on (strength every level, constitution every second, dexterity every third), item bonuses capped at 1.5 × level.
  - **Hit points and endurance:** class base hit points × level plus a constitution share; items add up to 4 × level. Endurance is 100.
  - **Specialisation:** spec points per level (multiplier 2.0, half-levels from 40), training at a Warrior trainer at the cost of the next level, specs capped at character level. Weapon styles unlock by spec level, not character level.
  - **Melee:**
    - weapon skill from class base, level, strength and spec;
    - damage per second capped by level;
    - armour factor and absorption by hit location;
    - quality and condition;
    - variance driven by the spec-to-target ratio, and a damage cap;
    - critical hits;
    - style damage from the growth rate, endurance cost by weapon speed, and the style's to-hit and defence bonuses;
    - a base 18% miss chance moved by the level difference;
    - evade, parry and block from abilities, spec and dexterity, then misses;
    - swing speed reduced by quickness, with a 1.5 s floor.
  - **Creatures:** hit points, weapon skill, damage, armour factor, absorption and swing speed from level and stats, the template's parry, evade and block, and their race's resists to crush, slash and thrust.
  - **Experience:**
    - the real experience table and creature experience values;
    - con colours;
    - the kill cap at 1.25 × a same-level kill;
    - a share for each player by damage done;
    - nothing for grey kills.
  - **Death:**
    - experience loss from level 6, part of it held in a gravestone that praying recovers;
    - constitution loss on release;
    - healers who restore it for money;
    - experience never falls below the start of the current level.
- **Equipment rules from the class.** A Warrior may use one- and two-handed swords, axes and hammers, shields with one-handed weapons only, and armour up to chain. **BREAKING:** today's "a warrior here fights one-handed" refusal goes, and two-handed weapons work.
- **Protocol (additive):**
  - a `stats` message (stats, spec points, specs, abilities, weapon skill, armour factor);
  - a con colour on each creature entity;
  - trainer and healer interactions (`ik:"trainer"`, `ik:"healer"`) answered with `trainer` and `healer`;
  - `train` and `restore` requests;
  - `pray`;
  - gravestones as entities;
  - death-penalty events.
- **Store:** characters keep base stats, spec levels, unspent points, death count, constitution lost and their gravestone. Existing characters are migrated to the default creation build and given the spec points their level has earned, unspent.
- **Client:** built-in panels for the character sheet, the trainer and the healer (the stock windows come later and fall back to these), and gravestones drawn as a marker.

Out of scope, each listed in the plan above:
- the base class at levels 1–5 (Viking, choosing Warrior at 5) and character creation choices;
- the Dwarf, Kobold and Valkyn archetypes at creation (their stats are defined here);
- realm abilities, player resists, buffs, magic, groups and the group experience bonus, the camp bonus, resurrection by players, RvR;
- active abilities the Warrior is granted (Protect, Intercept, Guard, Engage): they are listed as known but inactive;
- jewellery slots (neck, cloak, belt, bracers, rings);
- NPC behaviour changes (respawn, aggro, factions);
- the local OpenDAoC reference harness.

## Capabilities

### New Capabilities
- `rule-book`: how game rules are recorded (one spec per mechanic, our own words, sources and status), and the parity census that measures the server against the data.
- `rules-character-stats`: race and class definitions, creation points, level growth, item stat bonuses and caps, hit points, endurance and regeneration.
- `rules-specialization`: spec points per level, training at a trainer, spec caps, and what spec levels unlock.
- `rules-melee`: the melee attack from swing to damage: speed, defences, miss, hit location, weapon skill, armour, variance, caps, criticals, styles.
- `rules-creature-combat`: the combat numbers of creatures, from level, stats and template.
- `rules-experience`: the experience table, creature experience values, con colours, kill experience and its cap.
- `rules-death`: the death penalty: experience and constitution loss, gravestones, praying and healers.

### Modified Capabilities
- `character-progression`:
  - "Experience and levels" becomes the real table, and levelling grants stats and spec points instead of unlocking styles by level;
  - "What a Warrior may carry and wear" allows two-handed weapons;
  - "Equipment sets the numbers" adds item stat bonuses.
- `server-combat`:
  - "Death of a player" gains the penalty, the gravestone and constitution loss;
  - "Styles and defences come from data" gets its numbers from the rule book.
- `npc-interaction`: creatures that train or heal answer an interaction, as merchants do.

## Impact

- **Code:**
  - new `internal/server/rules` package and `cmd/parity`;
  - `internal/server/world` (`rules.go` shrinks to timing and layout constants; `combat.go`, `npc.go`, `player.go`, `inventory.go`, `world.go` call the rules package);
  - `internal/server/store` (schema migration);
  - `internal/importer` (item bonuses, quality and condition, spawn stats, trainer and healer kinds);
  - `internal/server/protocol`, `docs/protocol.md`;
  - `web/` (panels, gravestone marker, death messages);
  - `internal/server/gameserver` (creation validates a stat build).
- **Data:** `dbimport` must be re-run, because the world store gains item bonus columns and spawn stats. `players.db` migrates in place.
- **Docs:**
  - CLAUDE.md and the README gain the rule book convention and the sources list;
  - `docs/parity/` is new, and committed because it holds counts and names only.
- **Behaviour players will notice:**
  - level 1 Warriors have about a quarter of today's hit points (a level 1 Troll Warrior goes from 175 to about 40) and creatures hit harder at low levels;
  - styles need trained spec;
  - dying costs experience and constitution.

  The rule book says which numbers are provisional, so they can be revisited.
