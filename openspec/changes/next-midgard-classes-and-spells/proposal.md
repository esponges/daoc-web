## Why

`midgard-playable` made a Warrior playable with friends, but Warrior is the only class a player can be. The census says `Class spell types 0 of 73 implemented`, the SPELLS window is greyed ("Players cannot cast spells yet"), and every character is male. A 1.65 Midgard group needs a healer, a caster and more than one melee class, and the playtest in `midgard-playable` (findings 18.1) listed what felt unlike 1.65. This change is the next step toward a 1.65-faithful Midgard played with friends.

## What Changes

The change is cut to one slice a group can feel: five classes, one new mechanic (player spellcasting with power), and the fixes from the playtest. The other Midgard classes are the change after.

**Tier 1: fixes from the playtest** (cheap, and every session meets them)
- **Attack control:** pressing attack when already attacking no longer stops it silently; the bar shows the attack state. Attack at a distant target walks the character to it, as 1.65's auto-walk did. Both checked against the rule sources before they are decided.
- **Loot:** kills pay by creature level and kind, not coin only. The imported data has no loot generators, so the table is ours, marked provisional, and a source is searched for first.
- **Night spawns:** dusk brings night spawns out staggered instead of all at once, if a source shows 1.65 did; otherwise the finding is recorded as unresolved.
- **`equip` between pack slots:** the server rejects a pack-to-pack equip with a clear message.
- **Gender:** character creation lets the player choose male or female. The character stores it and every client draws it. This closes `midgard-playable` 17.2's same-race gender pair.

**Tier 2: player spellcasting**
- **Power:** a power pool and regeneration, in a new rule-book spec.
- **Spellbook and casting:** each class has spell lines from the imported data, spells learned by level, cast with range, cast time, interrupt by damage or movement, recast and power cost. The creature casting engine is reused where it fits.
- **Spell types, first slice:** direct damage, bolts, heals, damage over time, damage add, strength and constitution buffs, armour buffs, speed decrease, mesmerize and stun. These are the types the five classes use most.
- **SPELLS window and bar:** the stock SPELLS window lists the spellbook; spells go on the quickbar; a cast bar shows cast time.
- **Banking and repair:** both checked in play, because the last session did not exercise them.

**Tier 3: five classes**
- **Berserker and Savage:** Viking-base melee with their own styles and weapon specs. Styles are already imported; their class lists are the new data.
- **Healer, Shaman, Runemaster:** the three Midgard casters a group needs most (heals and buffs, hybrid support and damage, bolts and direct damage). Each gets stats, hit points, spec lines, trainers and the quickstart kit from the data.
- Classes of the same base share the base's trainers; advanced-class choice at level 5 stays out of this change.

**Tier 4: integration**
- Two clients play a group of a Warrior, a Healer and a Runemaster through the Vale of Mularn: heal, buff, kill, loot, die, pray. Findings go to the next change.

**Not in this change:** Bonedancer, Spiritmaster, Skald, Thane, Hunter, Shadowblade, Valkyrie, Seer, Mystic and the pets, charm and summon spells that go with them. Realm abilities, quests and crafting also wait.

## Capabilities

### New Capabilities
- `player-spellcasting`: casting by players: spellbook, quickbar, cast time, range, interrupt, recast, the first slice of spell types, and the SPELLS window.
- `midgard-classes`: the five classes' stats, hit points, specs, trainers and starting kit, and the gender choice at creation.
- `rules-power`: the power pool, its regeneration and what casting costs, in the rule book's form with sources and status.

### Modified Capabilities
- `loot-and-release`: kills pay items by creature level and kind, not only coin.
- `character-progression`: classes other than the Warrior level, train specs and gain power.

## Impact

- `internal/server/rules` (power, class table, spell effects), `internal/server/world` (`casting.go`, `player.go`, `catalog_spells.go`, `catalog_start.go`, `groups.go`), `internal/server/protocol` and `docs/protocol.md` (additive: spellbook, cast, power, gender).
- `cmd/dbimport` (class spell lines and class starts), `cmd/charconv` (female models for Midgard races), `cmd/parity` (class spell types counted).
- `web/` (`ui.js`, `hud.js`, `net.js`, `uiwin.js`: spells window, cast bar, power bar, attack state, gender choice in the creation page).
- `players.db` gets additive columns (gender, class, spellbook, power). The live server needs a backup and a restart, as in `midgard-playable` 4.1.
- No game data is committed; the community database stays imported at build time.
