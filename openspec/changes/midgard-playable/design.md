# Design

## Context

See proposal.md for why. The current state that shapes the approach:

- **One goroutine per zone.** Each zone runs its own tick (`world/zone.go`), and `World` holds what spans zones under `w.mu`. Moving between zones goes through `moveToZone`. Anything that spans zones (groups, tells, the clock) lives on `World`; anything that acts inside a fight (spells, effects, guards) lives in the zone tick.
- **NPCs hold one target, and it is a `*Player`.** `NPC.target` is a player, with no hate list, and `npcSwing(n, p)` only hits players. Guards fighting creatures, Protect moving anger, and "turns on whoever it hates most" all need more than that.
- **Movement is shared maths.** `world.Step` and `web/move.js` run the same fixed steps, checked against vectors that the Go tests write. Speed is `Walk` or `Run`, with no modifier.
- **No stored bind.** `bindFor` picks the nearest imported bind point in the player's zone, and falls back to `Catalog.StartFor`.
- **Starter kit:** `GiveStarterKit` picks the cheapest matching items. The data's `starterequipment` table is not imported.
- **The rule book:** each rule is a pure function in `internal/server/rules`, with a spec, a status, sources and a table test per scenario, registered for `cmd/parity`. `world` calls them.
- **The stock UI:** `uiwin.js` renders windows from converted layouts. The plain panels in `ui.js` are the fallback.
- **The imported data has what this change needs:**
  - start rows per race and class: the Troll row is at about (797583, 893718) in region 100;
  - `starterequipment`, with class lists such as `35;31;22;`;
  - `spell`, the `npctemplate.Spells` lists, `stylexspell` (almost all at chance 100) and item `ProcSpellID`/`ProcChance`;
  - `path`/`pathpoints`, with horse routes named `HS_<from>_<to>`;
  - 509 `DOL.GS.NightSpawn` mobs;
  - class types `GameStableMaster`, `GameVaultKeeper`, `Blacksmith` and `GameGuard`;
  - item `Condition`/`MaxCondition`;
  - Warrior abilities in `specxability`: Protect at career 7, 13 and 18, Intercept at 12, Guard at Shields 5, 10 and 15, and Engage at Shields 7.
- **The live server** (the user's) holds `world.db` open, so `dbimport` cannot replace it today.

## Goals / Non-Goals

**Goals:**
- Each tier leaves the server playable and its tests green, so the user can play between tiers.
- Every new number is a registered rule with a status. Anything provisional has a task to find a source.
- `players.db` changes are additive, with defaults, so old characters load unchanged.

**Non-Goals:**
- Player spellcasting. Spells here are cast by creatures, styles and items, and the effect code is written so that player casting can reuse it later.
- Line of sight for spells. There is no world line-of-sight model yet, so creatures cast at anything in range.
- Guild, broadcast-tab and reply chat.
- Horse art beyond what the install converts.

## Decisions

### 1. Staged re-import, swapped at start
`dbimport` writes `world.db.next` when it cannot take `world.db`. The server, before opening the store, checks the staged file: it opens, has the required tables, and has more than zero spawns. If so, it renames `world.db` to `world.db.prev` and the staged file into place. This is the only way to switch the live server without stopping it first, and it gives a one-step rollback.
- *Alternative:* the server reloads the world store live. Rejected: every zone, NPC and catalog pointer would need to be rebuilt under load, for no gain over a restart.
- The switch-over of the user's server is its own task. It asks the user first and copies `players.db` aside before the restart.

### 1a. Templates that replace mob values win (found during 2.1)
The danger at the Troll start was ours, not the data's. Nearly every mob (91,231 of 97,737) has an NPC template marked `ReplaceMobValues`. OpenDAoC-Core, read to understand how its data is meant to load, sets a mob from its row and then, for such a template, overwrites:
- the level, picked at random from the template's list each time the creature spawns;
- the name, aggro level and range, flags, speed and melee type;
- the model, size, equipment and class type.

Our importer ignored the mark. 8,557 rows (2,640 in Midgard) have a level outside their template's list, and 28,844 a different aggro level. For example, the giant bull frogs and arachite hatchlings 2,500 units from Galplen have rows at levels 25–29 and aggressive, but templates at level 1 and passive.

The importer now applies the template's name, levels, aggro, flags, speed and class type to the spawn, and stores the level list. The server picks a level from the list at each spawn and respawn. Model, size and equipment stay the row's: the client draws only converted models, and a template's model list has entries with a flag bit (`118;32886`) that is not yet understood. That is a later fix if it shows. Template level columns are also lists (`13;18;16`), which `parseRange` read as 0. `importer.LevelList` reads them now.

### 2. Safe start and the stored bind are decided at catalog load
**Source (task 2.1):** ZAM's "Starting Zones" table ([camelot.allakhazam.com/Start_Zones.html](https://camelot.allakhazam.com/Start_Zones.html)) gives each Midgard race and base class a start town. It is from the Shrouded Isles and Trials of Atlantis era (Valkyn and Frostalf are listed), which includes 1.65. It agrees with the imported start rows: Troll Vikings in Galplen (Myrkwood Forest), Norse Vikings in Vasudheim (East Svealand), Dwarf Vikings in Mularn, Kobold Vikings in Fort Atla (Gotar). The Troll start is therefore the classic one, and with decision 1a it is safe. The unsafe-start rule stays as a safety net, and it reports any start it replaces.

`Catalog` already loads starts and spawns. After loading, it tests each start against aggressive spawns of level 6 or more within 1,500 units, records the unsafe ones with their worst creature, and maps each to the realm's safe start. `StartFor` returns the replacement. The rule is a registered function (`rules.UnsafeStart`) with a provisional status, so it is counted and tested on synthetic spawns.

Bind becomes four columns on the character (zone, x, y, z, plus yaw). A null bind means a character from before this change: on first login it is set to the nearest bind point, and then the "moved once" check runs. `/bind` uses the imported bind points' radius.
- *Alternative:* make the dangerous creatures passive near the start, as `stock-ui-followups` had proposed. Rejected: it changes creature behaviour the data gives, and the danger would still be there for a level 2 character a few steps away.

### 3. Starting equipment from the table
The importer splits each `starterequipment` row's class list into a `start_equip(class, item, ord)` table. `GiveStarterKit` becomes a lookup: equip each item in its slot if the slot is free and `rules` says the class may use it, otherwise put it in the pack. Its cheapest-item search is deleted.

The levelling-gear check is a section of `cmd/parity`. It reads `merchantitem` joined to items and to Midgard merchants, by level band and kind. It goes in the census because it is a statement about the data, not about the server.

### 4. A combatant interface and hate tables (before any Tier 2 work)
NPCs get `hate map[combatantID]float64` in place of `target *Player`, where a combatant is a player or an NPC. Damage adds hate, Protect adds hate to the protector, and the NPC fights the highest-hate combatant in its zone. Packmates calling for help copy the hate, as today's calling does with the target. `npcSwing` takes a combatant. Players hitting NPCs are unchanged.

This refactor comes first in Tier 2, and the existing sim tests are its safety net. Guards, Protect and group fights all depend on it.
- *Alternative:* special-case guards with an NPC target field. Rejected: Protect and the group experience split also need to know who has done what, so a second ad-hoc field would follow.

### 5. Effects live on the combatant and are ticked by the zone
An `Effect` has a spell, caster, start, end, next tick and kind. Each player and NPC carries a small list. The zone tick advances it: damage-over-time ticks, ending, mesmerize breaking on damage. The formulas (hit chance, variance, resist, tick count, drain, heal) are in `rules/spells.go`. The bookkeeping is in `world/effects.go`.

Players are sent their own effects (name, a kind for the icon, time left). Icons are drawn by effect kind, not by the imported icon number, following the rule in CLAUDE.md.

### 6. Slows and holds go through the shared movement step
`world.Step` and `move.js` get a speed factor on the body: 1 normally, the strongest slow otherwise, and 0 while stunned or mesmerized. The server sends the owner their factor whenever it changes, so prediction applies the same factor at the same step. New shared vectors cover the factor, and both tests check them.
- *Alternative:* the server alone slows and the client corrects. Rejected: every slowed step would snap back, which is the latency hiding we already solved.

### 7. Creature casting is a decision in the NPC's combat step
At each combat decision:
1. heal if a heal is ready and someone is below 75%;
2. else cast a ready hostile spell in range;
3. else swing or close in.

A cast is a state with an end time. Any melee attack on the caster ends it unless the spell is uninterruptible. Both the decision and the interrupt are rules functions, with provisional status where the specs say so. Clients get a `cast` event (caster, spell name, target, duration) and print the line. No cast bar is drawn for another's cast.

### 8. Procs ride on the existing hit path
After `rules` settles a blow as landed:
- the style's `stylexspell` rows roll (by class 0 or the attacker's class);
- then the weapon proc;
- then, for a creature, its offensive proc.

Each one cast skips the hit roll and applies through decision 5.

### 9. Groups and chat live on World
`World` keeps the groups (id, leader, ordered members) and an online-name index, both under `w.mu`. Group state goes to each member once a second, from that member's zone tick, built from a snapshot of the other members (health, endurance, zone, and distance when in the same zone). Tells and `/g` look up recipients through the index and send through each player's existing send queue. The experience split runs in the zone where the kill happened. Members in other zones are out of range by definition, so the split only needs that zone's players.

### 10. The clock is a function of real time
Game seconds = (Unix seconds × 24) mod 86,400, so one game day is 60 real minutes. It needs no state, survives restarts, and every zone computes the same value. Each zone checks once a second whether day or night has flipped, and spawns or retires its night-only spawns. A retiring spawn in combat is marked to leave when its fight ends. Clients get the clock on entering the world and once a minute after, and interpolate between.

### 11. Horse routes are a server-driven rider state
A ticket handed to a stable master looks up its route. Which column of the ticket names its path is settled by a spike over the data (task 11.1), and the link is recorded in the README; it is not guessed.

A rider has a `ride` state: it ignores inputs, follows the path's points at each point's speed (or the default), and crosses zone borders through `moveToZone` with its place on the route kept. Creatures skip riders when choosing targets. Others see the rider through a `ride` entity field. The horse model is converted if the install has one; otherwise the rider is drawn running.

### 12. The vault reuses inventory rows
Vault slots are inventory rows numbered 1000–1039, so a move between pack and vault is one row update in the existing transaction code. The pack, the paper doll and loot ignore that range. A vault keeper's interaction opens it. Moves are refused out of reach.
- *Alternative:* a vault table. Rejected: swaps between tables need two-table transactions for no benefit.

### 13. Item condition is a column on inventory rows
Condition is stored per item in hundredths of a percent. Old rows default to 100%. A new item takes the template's `Condition`/`MaxCondition` when it has one. The wear and effectiveness rules are in `rules/condition.go`. Combat applies effectiveness where it already reads DPS and armour factor. A blacksmith's interaction opens a repair panel: drop an item on it to see the price, then confirm.

### 14. Not-classic NPC kinds are a sourced table
A Go table maps class type to reason and source, and it starts with the list in the `npc-interaction` spec. `npcKind` returns `notclassic` for these. `cmd/parity` reads the same table and prints those kinds in their own column. Enchanters, rechargers and craft masters stay plain NPCs and are not marked: crafting is a later change, and whether enchanters were in 1.65 is not settled.

### 15. Stock windows follow the existing renderer
`StatsAttributes`, `Stats_Spec_Abil`, `StatsGroup`, `MiniGroup` and `new_group` are rendered through `uiwin.js`, filled from data the client already has (`net.stats`) or from the new group messages. Each has a plain-panel fallback in `ui.js`. Fields the server does not send stay empty.

### 16. What came in from the closed changes is done by their own designs
- The interior fit (Trollheim, Tuscaran Glacier, Nisse's Lair) follows §3 of `archive/2026-10-19-jordheim-and-midgard-dungeons/design.md`.
- NPC faces start from the converter finding in `archive/2026-10-19-npc-equipment/tasks.md` 4.5.
- The browser checks follow the notes already recorded in those archives.

## Risks / Trade-offs

- **The hate-table refactor touches all NPC behaviour.** → It goes first in Tier 2, alone, with the sim and real-zone tests as the guard and a browser pass before anything builds on it.
- **Many numbers are provisional:** group split, casting cadence, ability values, wear and repair. → Each is a registered rule with status `provisional`, visible in `cmd/parity`, and has a sourcing task. The parity test fails if a rule's status weakens.
- **Slowing the shared movement step could break prediction.** → New vectors, and both test suites must agree to zero difference before the client uses the factor.
- **Night spawns appearing at once could spike a zone.** → They spawn over the first game minute of night, not all in one tick.
- **The 60-minute day is reference, not documented.** → The day length is one constant with a sourcing task. Changing it changes no other design.
- **The ticket-to-path link might not be in the data.** → The spike records it. If there is no link, routes are chosen by the stable's town and the ticket's destination name, recorded as provisional.
- **This is a large change.** → It is phased in tiers. Each tier ends with a playable server, a commit when the user asks, and, for Tier 1, the live switch-over.
- **Spells without line of sight** let creatures cast through walls in dungeons. → Recorded in the README's limitations, and left for a later change with a line-of-sight model.

## Migration Plan

1. **Tier 1 code.** `players.db` gains bind columns and an item condition column, both nullable or defaulted, and old rows load unchanged.
2. **The user's server.** With the user's go-ahead:
   - copy `players.db` to a dated backup;
   - run `dbimport`, which stages `world.db.next`;
   - restart the server;
   - check the load report (replaced starts, kit counts).
3. **Rollback:** stop, put `world.db.prev` back as `world.db`, restore the `players.db` backup, and restart the previous build.
4. Tiers 2 and 3 repeat step 2 when they change the import.

## Open Questions

- **The 1.65 day length and night hours.** Reference until sourced; a single constant.
- ~~**The 1.65 start towns per race and class.**~~ Answered by ZAM's table (decision 2): the data's starts agree with it.
- **Template models.** Whether to take a replacing template's model list, and what the flag bit in `32886` means.
- **Intercept's duration, Protect's share, and Engage's endurance per block.** Provisional until sourced; numbers only.
- **The unit of `LifeDrainReturn`.** Settled by checking several drain spells' values in the data.
