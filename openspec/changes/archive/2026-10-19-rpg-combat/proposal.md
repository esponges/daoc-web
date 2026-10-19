> **Closed 2026-10-19, not built as proposed.** Superseded by `rules-parity-warrior`, which gave the Warrior its styles, defences, endurance and the 1.65 rules from the imported data and the rule book instead of an authored ruleset. No specs were synced from this proposal.

# Proposal

## Why

Combat today is auto-attack and a shield block with this project's own numbers. DAoC's combat is defined by classes and styles — named special attacks that chain off openings, reward position and cost endurance — and that is what makes it recognisably DAoC. The client ships the whole presentation of styles (583 named styles with icons, animations, durations and trail colours, plus per-race style, parry, evade and block clips) but none of the rules, which live on the server. This change uses the client's presentation data as-is and authors a small, clearly labelled ruleset of our own, starting with the Warrior class the player's troll already plays.

## What Changes

- Presentation from the game data:
  - styles from `styles.csv`: name, icon, animation, duration, build-up, trail colour
  - per-race style animations from `canims.csv` (the "Style 1HC 1"… columns), converted by `animconv` alongside the existing combat clips
  - parry, evade and block clips per weapon type from `canims.csv`
  - weapon trails in the style's colour
- Authored rules (this project's, labelled as such, modelled on DAoC's documented mechanics, not copied from any server database):
  - a `Warrior` class definition: weapons and armour allowed, style lines (hammer, axe, sword, shield, parry), and which styles each line grants
  - style requirements: anytime, opening after a block, parry or evade, positional (side, behind), and follow-ups in a chain
  - endurance: styles cost endurance, which regenerates
  - defence: parry and evade in addition to block, with their animations
- UI: a style hotbar with the game's icons, key bindings (for example 1–9), showing which styles are currently available, and combat log entries naming the style used.
- NPCs keep simple auto-attack, but can parry, evade and block.

## Capabilities

### New Capabilities
- `combat-styles`: style presentation from game data and the rules for when a style can be used and what it does.
- `character-classes`: authored class definitions (starting with Warrior): allowed equipment and the style lines and styles they grant.
- `melee-combat`: auto-attack, hit resolution, block, parry, evade, endurance and the combat log, generalising today's combat.

### Modified Capabilities

## Impact

- `cmd/animconv`: style, parry and evade roles from `canims.csv`; `internal/gamedata`: `styles.csv` and `icons.csv` readers; style icons converted from the icon archives.
- New authored data file(s) for classes and style rules, in the repo (our own content, MIT).
- `web/combat.js` grows considerably; new hotbar UI; `combat.test.mjs` extended for openings, positionals and endurance.
- Licensing: no server emulator databases are imported; all rule numbers are ours.
- Independent of `world-census` and `cities-and-dungeons`.
