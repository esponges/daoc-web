# Proposal: hibernia-realm

## Why

Hibernia is the third realm, and classic RvR is three-sided. Relics and keeps assume three realms, and the frontier balance only works with all three. Like Albion, Hibernia's zones convert but are not served and have not been looked at.

1.65 Hibernia has thirteen classes:

| kind | classes |
| --- | --- |
| melee | Hero, Champion, Blademaster |
| healers and support | Druid, Warden, Bard |
| stealth and archery | Nightshade, Ranger |
| casters | Enchanter, Eldritch, Mentalist, Animist |
| hybrid | Valewalker |

Its regions are 200 (Hibernia) and 181 (the Shrouded Isles: Hy Brasil).

This change depends on `albion-realm`, which makes the realm a parameter. Here Hibernia fills it in.

## What Changes

- **World:** Hibernia's zones, Tir na Nog and its dungeons are served with their population. Zone travel, horse routes and start towns are added, and every zone is checked by eye.
- **Classes:** all thirteen. These are the new mechanics:
  - the Animist's turrets: stationary pets placed at a ground target, which needs the ground target from `finish-midgard-classes`;
  - the Bard's songs, using the shared song code;
  - the Champion's debuff-strike spells;
  - the Mentalist's power and health transfer;
  - the Enchanter's bomber pets;
  - the Valewalker's scythe line;
  - the Blademaster's dual wield (Celtic Dual), using the shared left-hand code.
- **Census:** a Hibernia section.
- **Gear** to level 50.

## Capabilities

### New Capabilities
- `hibernia-classes`: the thirteen classes, their trainers and lines, and turrets.

### Modified Capabilities
- `world-data-import`: the Hibernia regions are imported.
- `zone-travel`: Hibernia's borders, dungeon exits and horse routes.

## Impact

- **Import and converters:** Hibernia regions; race figures (Celt, Firbolg, Elf, Lurikeen, Sylvan) and creature models.
- **Server:** class entries; turrets as stationary pets; bomber pets.
- **Census:** a per-realm section.
- **Client:** Hibernia in creation; the realm skin if it differs.
