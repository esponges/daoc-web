# Proposal: albion-realm

## Why

Midgard is the only playable realm. Albion's zones, cities and dungeons convert (`worldconv`) and the viewer opens them, but they have only been measured, not looked at, and the server neither imports nor serves their population. Realm-versus-realm play needs at least two realms with players in them. Since this is a friends' build, each realm must be playable from level 1, so friends can pick sides.

1.65 Albion has fourteen classes:

| kind | classes |
| --- | --- |
| tanks | Armsman, Mercenary, Paladin, Reaver |
| healers | Cleric, Friar |
| stealth and archery | Infiltrator, Minstrel, Scout |
| casters | Sorcerer, Theurgist, Wizard, Cabalist, Necromancer |

Its regions are 1 (Albion) and 51 (the Shrouded Isles: Avalon).

## What Changes

- **World:**
  - Albion's outdoor zones, Camelot, and its dungeons are served, with their population imported from the community database (regions 1, 51 and the Albion dungeon and city regions);
  - zone travel, horse routes, bind points and start towns come from the start table (ZAM);
  - each zone is checked by eye, as Midgard's were: terrain, placements, interiors and exits.
- **Classes:** all fourteen are made playable with their trainers, specializations, styles, spells and abilities. Two are new mechanics:
  - the Necromancer's shade form: the Necromancer becomes a shade while its pet fights, and damage to the pet is the player's;
  - the Minstrel's songs, which reuse the Skald's song code.
- **Realm separation:**
  - a character belongs to one realm, chosen at creation;
  - chat, groups, guilds, trade and tells are within a realm;
  - players of another realm show as enemies, though they cannot meet until the frontiers or battlegrounds.
- **The census** grows an Albion section with the same counts as Midgard's.
- **Merchants and gear** to level 50, filling bands the community data leaves empty by the same added-stock rule (`internal/importer/addedstock.go`).

## Capabilities

### New Capabilities
- `albion-classes`: the fourteen classes, their trainers, lines and the class-specific mechanics (shade form).
- `realm-membership`: a character's realm, and what it allows (chat, groups, trade, guilds within the realm; hostility across).

### Modified Capabilities
- `world-data-import`: the Albion regions are imported.
- `zone-travel`: Albion's zone borders, dungeon exits and horse routes.

## Impact

- **Import:** `cmd/dbimport` region list and Albion start table; re-import.
- **Converters:** `worldconv` (already run) plus `charconv` and `animconv` for Albion races and creature models; `uiconv` covers the realm skin if the install's UI differs by realm.
- **Server:** catalogue loading per realm; class rules (`rules-*` updates for each class line); shade form; realm checks on chat, groups and trade.
- **Census:** per-realm sections.
- **Client:** creation lists realms, races and classes; the realm colour on name plates.
