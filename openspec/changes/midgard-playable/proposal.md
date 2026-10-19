## Why

`rules-parity-warrior` made the Warrior follow the 1.65 rules, but Midgard still isn't a server you can test with friends:

- A new Troll appears among level 26 creatures.
- The starter kit is our own guess.
- There are no groups.
- Creatures never cast.
- You walk everywhere.
- Guards stand idle.
- Items never wear, and there is no bank.
- The menu bar's character windows are greyed.

This change does the pending work that lets a Warrior, alone or with friends, level through Midgard as on a 1.65 live server. It is now the only open change: the unfinished parts of `stock-ui-followups`, `npc-equipment`, `npc-gear-mapping` and `jordheim-and-midgard-dungeons` moved here when those changes were closed.

## What Changes

**Tier 1: unblock testing**
- **Re-import while running:** `dbimport` can stage a new world store while the server holds `world.db`. The server swaps it in on its next start. Switching the live server to the new rules becomes a restart, not a shutdown dance.
- **Safe start:**
  - The data's Troll start lies in Myrkwood Forest, beside level 26 creatures.
  - A start whose surroundings are far above level 1 is replaced by a safe start for the realm, chosen by a rule and reported at load. The data is not edited.
  - The 1.65 start towns are looked for in a source. If one is found, it replaces the rule's choice.
  - Characters get a stored bind point and `/bind` at a bindstone, as in 1.65. Today a dead player returns to the nearest bindstone in their zone.
- **Classic starting equipment:** a new character gets the items the data's `starterequipment` table gives their class (for a Warrior: bronze stelskodd armour, a training axe and a small training shield). This replaces our cheapest-item guess.
- **Gear for levelling:** a data check that Midgard merchants sell Warrior weapons and armour for every level band from 1 to 50. Gaps are reported, not invented.

**Tier 2: a live server with friends**
- **Groups:**
  - Invite, accept, leave, disband, and a leader.
  - Members see each other's health in the group window.
  - Group chat and `/tell` work.
  - Members within range share experience by the classic group rule.
- **Creature spells:**
  - Creatures whose templates list spells cast them: direct damage, bolts, damage over time, slows (speed decrease), mesmerize, stun, life drain and heals. Together these are most of the creature spell uses the census counts.
  - Spells have range, cast time, interrupts and recast, and are resisted by the target's resists.
  - Styles and weapons with a proc spell cast it on hit.
- **Horse routes:** stable masters sell tickets, and handing one in rides the data's route (`HS_<from>_<to>` paths) to the next town.
- **Classic-only NPC kinds:**
  - NPC kinds that 1.65 did not have stay inert and say so. Teleporters and hasteners are among them unless a source shows them in 1.65.
  - The census marks such kinds "not classic" instead of "missing".
- **Guards:** town guards attack aggressive creatures that come into their range near players of their realm.
- **Warrior abilities:** Protect, Guard, Intercept and Engage work as 1.65 had them.
- **Vault:** vault keepers store items per character.
- **Item wear and repair:** weapons and armour lose condition in combat, and blacksmiths repair them for money.
- **Day and night:**
  - A game clock that every client shares.
  - The data's night-only spawns (`NightSpawn`) appear at dusk and leave at dawn.
  - The stock clock shows the game time.

**Tier 3: the game's own windows, and checks left over**
- **Character windows:** ATTRIB., SKILLS and STYLES open stock windows built from their layouts and filled with what the server holds. GROUP opens the group window. SPELLS, JOURNAL, MAP and WARMAP stay greyed, with a tooltip saying why.
- **Compass and clock:** the compass's north is checked against a zone exit, and the clock shows the game time.
- **Seen end to end:** a real kill's loot through the stock loot window, and a real death through the Release window.
- **Icons:** one bounded search for a real map from the imported icon numbers to this install's sheets. If none is found, the kind-based pictures stay.
- **Looks:**
  - NPC faces.
  - The remaining checks that races, genders and worn gear look right to every client.
- **Interiors:** a better fit for Trollheim, Tuscaran Glacier and Nisse's Lair, and a tour of all 13 Midgard interiors.

**Left for later changes:** the other eleven Midgard classes, player spellcasting, ranged combat and stealth, Dwarf, Kobold and Valkyn outfits, quests, crafting, guilds, and realm vs realm.

## Capabilities

### New Capabilities
- `starting-zone-safety`: where a new character appears and what may attack it there.
- `starting-equipment`: what a new character is given, and the check that levelling gear exists for sale.
- `groups`: forming and leaving groups, the leader, what members see of each other.
- `chat-channels`: tell and group chat alongside say and broadcast.
- `rules-spell-effects`: what each creature spell type does to its target, its resist and its duration.
- `rules-creature-casting`: when and how a creature casts, and what interrupts it.
- `rules-abilities`: Protect, Guard, Intercept and Engage.
- `rules-item-condition`: condition loss in combat and the price of repair.
- `horse-routes`: stable masters, tickets and riding a route.
- `vault-storage`: the vault keeper and what a vault holds.
- `world-time`: the game clock, and spawns that keep to night.
- `character-sheet`: the attributes, skills and styles windows, and what the server tells them.
- `world-orientation`: the compass's agreement with the map.

### Modified Capabilities
- `rules-experience`: a group's kill is shared among members by the group rule. The damage-share rule now applies between groups and solo players.
- `rules-melee`: a style or weapon with a proc spell casts it on hit.
- `npc-interaction`:
  - New NPC kinds: stable master, vault keeper, blacksmith and guard.
  - Kinds 1.65 did not have answer that they are not part of this world.
- `server-npcs`: guards defend their realm against aggressive creatures.
- `world-data-import`: a re-import can be staged while the server runs. Spells, creature spell lists, procs, paths, start equipment and the night-spawn mark are imported.
- `hud-chrome`: menu buttons whose windows now exist are enabled, and the rest say why they are greyed. The clock shows the game time.

## Impact

- **Server:**
  - `internal/server/rules`: new rule files and table tests for spells, casting, abilities, condition and group experience.
  - `internal/server/world`: groups, casting, guards, routes, vault, repair, the clock, and the start and kit changes.
  - `internal/server/store`: vault rows and item condition per stored item. Groups live in memory only.
  - `internal/importer`: spells, template spells, style procs, paths and points, start equipment, the night-spawn mark, and staged output.
- **Protocol:** new messages and fields only, documented in `docs/protocol.md`: attributes (owner only), group, tell, cast bars, effects, vault, repair, ride, clock.
- **Client:**
  - Stock windows: `StatsAttributes`, `Stats_Spec_Abil`, `StatsGroup`, `new_group` and `MiniGroup`, each with a plain fallback.
  - A cast bar and effect icons.
  - Vault and repair panels, and riding.
  - NPC faces in `charconv`.
- **Census:** `cmd/parity` counts spell types, abilities and NPC kinds as implemented, and adds a "not classic" mark.
- **Operations:** the user's live server switches over through the staged re-import. `players.db` is backed up first, and the switch is made only with the user's go-ahead.
- **Evidence kept from the closed changes:** `archive/2026-10-19-stock-ui-followups/design.md` (spawn danger, icons, compass) and `archive/2026-10-19-npc-gear-mapping/tasks.md` (what has been seen of gear in the browser).
