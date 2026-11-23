# Proposal: scripted-creatures

## Why

The parity census counts 110 creature kinds in Midgard's data. 23 are implemented and 7 are marked not classic. The rest are left out of the world entirely: the server skips any placement whose kind it does not know. Most of them are named monsters and their helpers, which give a zone its landmarks and camps.

The biggest gaps by placement count:

| kind | placements |
| --- | --- |
| Myling | 194 |
| NightSpawn / DaySpawn | 72 |
| ShiveringPresence | 51 |
| QueenMajorAdd | 22 |
| Enchanter | 14 |
| StripeAdd, ZritZritAdd | 13 each |
| OonaUndeadSoldier | 9 |
| RainbowSprite (four colours) | 28 |

There are also about 40 single named creatures, each with a script of its own:

- Stripe, ZritZrit, Snarls, Oona and QueenMajor;
- Goretooth, Shredclaw and Steinolf;
- the Wreth of Winter set;
- GlacierGiant, GiantLemer, RaumarikRevenant, Mortufoghus;
- and others.

Service NPCs are also missing:

- Recharger: recharges an item's charges;
- Enchanter: enchants an item;
- NameRegistrar and RoyalTreasuryClerk.

The keep guard kinds belong to `frontiers-and-keeps`, the crafting masters to `crafting` and the GuildRegistrar and EmblemNPC to `player-guilds`.

## What Changes

- **A behaviour layer for creatures.** A creature kind can add behaviours on top of the standard NPC:
  - time-of-day presence;
  - adds that spawn at health thresholds;
  - fleeing and calling for help;
  - one or more scripted abilities on a timer;
  - shape or colour cycling.

  Each scripted kind is described as data, a table of behaviours, not as one-off code where avoidable.
- **The common kinds:**
  - Myling;
  - NightSpawn and DaySpawn, folded into the existing night-only presence;
  - ShiveringPresence;
  - RainbowSprite;
  - the add kinds of the named bosses.
- **The named creatures.** Each is implemented from what its 1.65 behaviour is known to be, with a source in the rule book, or is spawned as a plain creature with its template and marked "behaviour unsourced" in the census.
- **Services:**
  - Recharger, for items with charges;
  - Enchanter, which adds the item bonus enchanting gave before crafting covered it, if 1.65 had it; otherwise marked not classic;
  - NameRegistrar and RoyalTreasuryClerk, after checking whether they are classic.
- **Left out:** the test dummies (DPSDummy, HitbackDummy, HealDummy), AtlasTrainer, OFMerchantHome, EffectNPC and the SI teleporters. They are marked not classic with reasons.

## Capabilities

### New Capabilities
- `creature-behaviours`: time-of-day presence, adds, fleeing and calling for help, timed scripted abilities, and how a creature kind maps to a set of them.

### Modified Capabilities
- `server-npcs`: every creature kind is either spawned or recorded as not classic, and service NPCs (recharger, enchanter) serve.

## Impact

- **Server:** `world/npc.go` and `world/zone.go`, plus a new `world/behaviours.go` with its tests.
- **Import:** `internal/importer` keeps the kind name and the script parameters it can read. `world.db` is re-imported.
- **Census:** `cmd/parity` creature-kind table and not-classic reasons.
- **Converters:** the charconv recipe covers the models of these kinds.
