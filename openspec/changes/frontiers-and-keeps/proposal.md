# Proposal: frontiers-and-keeps

## Why

The classic frontiers are where 1.65 RvR happened, and none of it exists here.

- **The frontier zones:**
  - Midgard's Odin's Gate, Jamtland Mountains, Uppland, Yggdra Forest and others;
  - Albion's Pennine Mountains, Forest Sauvage, Snowdonia and others;
  - Hibernia's Breifine, Cruachan Gorge, Emain Macha and others.
- **The keeps:** 79 keeps in Midgard's data alone, with 432 keep doors.
- **Darkness Falls:** a dungeon whose entry follows which realm holds the most keeps.

The census counts these keeps and doors at zero. The stock layout has the `Keep` window, the keep upgrade and claim buttons, the siege controls (`SiegeAim`, `SiegeFire`, `SiegeMove`, `SiegeRepair`, `SiegeSalvage`) and the war map with its relic selects.

This change builds on `keep-core` from `battlegrounds` and the RvR rules from `realm-versus-realm-combat`.

## What Changes

- **Frontier zones:** served as RvR areas with their population, their border keeps and the portal NPCs between the home zones and the frontier.
- **All frontier keeps and towers:** doors, guards, lords and ownership from the data, at their starting owners.
- **Claiming:**
  - a guild with a group in the keep claims a captured keep;
  - the guild's emblem shows on its banners;
  - upgrading the keep's level costs bounty points or gold (whichever 1.65 used, task 1.1) and takes time, raising guard and door strength;
  - releasing the keep is allowed.
- **Siege:**
  - rams, trebuchets, catapults and ballistas, bought from siege merchants or crafted;
  - placed on a ground target (ramps on the door);
  - aimed and fired, damaging doors and walls, or players with the anti-personnel ammunition;
  - repaired with woodworking;
  - boiling oil on hookpoints.
- **Keep repair:** owners with woodworking repair doors and walls.
- **Teleport between keeps:** owners use the keep portal stones (`keepdoorteleport`) to move between connected keeps.
- **Darkness Falls:** entry is open to the realm holding the most keeps, checked when keeps change hands.
- **The war map** shows keep ownership and fights in progress (the stock war map).
- **Climb Walls:** the Shadowblade (and the other realms' assassins) can climb keep walls at climb points, from `finish-midgard-classes`.

## Capabilities

### New Capabilities
- `frontiers`: the frontier zones as RvR areas, the border keeps and the portals in and out.
- `keep-warfare`: claiming, upgrades, siege engines, repair, hookpoints, keep teleports and wall climbing.
- `darkness-falls`: entry by keep ownership.
- `war-map`: keep ownership and fights in progress on the stock war map.

### Modified Capabilities
- (none; `keep-core` is introduced by `battlegrounds`)

## Impact

- **Import:** all keeps, components, hookpoints and door teleports for the three realms.
- **Converters:** siege engine figures.
- **Server:** `world/siege.go` and `world/claims.go`; Darkness Falls access.
- **Client:** siege controls (the stock Siege commands), the war map, keep upgrade and claim buttons, and banners.
- **Census:** keeps and keep doors implemented.
