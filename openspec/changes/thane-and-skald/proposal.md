## Why

Six of Midgard's twelve classes can be played. The Thane and the Skald are the two Viking hybrids that every 1.65 Midgard group had, and they are the next cheapest to add: their melee (styles, Shield, Guard, Intercept, Protect, Engage, Evade) already works, and most of their spells use types the server already applies. A Thane's spell line, Stormcalling, uses only direct damage, the strength/constitution buff and damage add, but 12 of its 44 spells hit an area (Thunder Shout around the caster, Mjollnir around the target), and the server has no area spells. Skald songs and chants pulse over the group while they are on, and the server has no pulsing spells either. A Thane or Skald added without those two would be missing what the class is known for.

## What Changes

- **Two classes.** Thane (class 21; Norseman, Troll, Dwarf) and Skald (class 24; Norseman, Troll, Dwarf, Kobold) join the class table with stats, hit points, weapon skill base, spec multiplier, races and casting stat (piety for the Thane, charisma for the Skald). Their races come from the data's start table, which agrees with ZAM's starting-zones table. Both have base class Viking, so the Viking trainers teach them as well as their own Thane and Skald trainers, and the census then counts those trainer kinds. The numbers come from OpenDAoC-Core (the only source) and are provisional.
- **Area spells.** The server reads each spell's `Radius`, which is already imported into `world.db` but not loaded. A hostile spell with a radius hits every creature that can be fought inside the radius: around the caster when its range is 0 (Thunder Shout), or around the target (Mjollnir). Each creature hit rolls its own hit and resist. Players are never hit, because there is no player-versus-player combat. Damage does not fall off toward the edge (the community servers differ, so the rule is provisional).
- **Pulsing songs and chants.** The importer adds the spell's `Pulse` and `PulsePower`. Casting a pulsing spell turns it on: it re-applies to the caster's group every `Frequency` tenths of a second, and its short effect lapses on a member who leaves range. Casting it again turns it off. Only one song is on at a time (the sources agree), and death turns it off; a held caster skips pulses.
- **Four new effect types, which together are eight spell types.** Speed enhancement (Songs of Travel), health regeneration (Songs of Rest), and the six resist buffs (body, spirit, energy, heat, cold, matter). Speed enhancement raises the speed factor that movement uses. Go `world/move.go` and JS `web/move.js` must agree on a factor above 1, and a shared vector checks it. The speed buff does not take hold in combat and ends when its holder enters combat. A resist buff adds to the resist a spell meets. The strongest buff of each kind counts.
- **Census and rule book.** The rule-book entries for the new effects, the area rule and the pulse rule are each registered with a status and table tests. `docs/parity` is regenerated: class spell types go from 13 of 73 to 21, and creature kinds gain the two trainers. The README counts and class list are updated. The Thane ability "Call of a thousand storms" stays counted as not implemented: no source was found for it.
- **Client.** The spell bar and effects list show a song that is on, plus the new effect kinds. Character creation lists Thane and Skald for the races that may take them.
- **Not in scope:** the other four classes (Hunter, Shadowblade, Bonedancer, Spiritmaster), debuffs, instruments (a Midgard Skald needs none), realm abilities (so "Call of a thousand storms"), and area spells for creatures. Area-capable creature spells keep their current single-target behaviour until a change of their own, which this change records in the findings.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `midgard-classes`: Thane and Skald become playable, with their races, trainers and spell lines.
- `rules-character-stats`: the class table gains the Thane and Skald definitions.
- `player-spellcasting`: area spells, pulsing songs and chants, and the speed, regeneration and resist buff types join what a player can cast.
- `rules-spell-effects`: how speed enhancement, the health regeneration buff, resist buffs and area damage work, with sources and statuses.

## Impact

- `internal/server/rules`: `character.go` (two classes), `spells.go` (speed, regeneration, resist and area rules, registry), with table tests.
- `internal/importer/world.go` and `cmd/dbimport`: `pulse` and `pulse_power` spell columns, additive. `world.db` is re-imported.
- `internal/server/world`: `catalog_spells.go` (radius, pulse), `pcasting.go` (area targets, song toggle and pulse), `effects.go` (new effect kinds, resist and regeneration), `move.go`/`player.go` (a speed factor above 1).
- `web/move.js` and the shared movement vectors; `web/spells.js` and `web/hud.js` (a song shown as on, new effect icons); the creation page lists the new classes from the server.
- `docs/protocol.md`: additive. New effect kind values, plus a field marking a pulsing spell in the book and whether it is on.
- `cmd/parity`, `docs/parity`, README, CLAUDE.md.
- Live server: back up `players.db`, re-import, rebuild and restart, as in earlier changes. Existing characters are unaffected.
