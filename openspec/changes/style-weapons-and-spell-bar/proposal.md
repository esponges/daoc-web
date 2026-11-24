# Proposal: style-weapons-and-spell-bar

## Why

A player reported on 2026-11-24, playing a level 20 Savage: "I can't find melee styles in the UI,
therefore can't place them in the action bar", and "cast spells are marked as red (unusable) even
after the time has passed". Checking it:

- **Two style weapon codes are never matched.** The imported style table's `weapon_req` is an
  object type, except two special codes: 1000, a style that needs a weapon in each hand (all 120
  Hand to Hand, Left Axe, Dual Wield and Celtic Dual styles), and 1001, a style for any weapon (the
  45 Critical Strike styles). `computeBar` compared them to the held weapon's object type, so they
  never matched: the Savage's Hand to Hand, the Berserker's and Shadowblade's Left Axe and the
  Shadowblade's Critical Strike were always "learned but cannot use". The styles window listed them
  greyed with "(needs Hand to Hand)", which reads as a spec they already have, and the style bar
  had nothing to drag onto.
- **The spell bar compared a Savage's spells with power.** Savagery spells are paid in hit points
  (`hp: true`, `pw` is the hit points). The stock spell quickbar shaded a slot red when power was
  below `pw`; a Savage has no power, so every Savagery slot was red for good. The plain bar already
  tested hit points.
- **The recast is invisible.** Savagery spells have a 5 second recast. The server refuses a cast
  in that time ("not ready yet"), but the bar does not show it.

## What changes

- `rules-melee` gains "A style needs its weapons": what the held weapons must be for each kind of
  style requirement, as a rule (`rules.StyleWeaponOK`) with table tests and a registry entry.
- The server builds the style bar with it, and each locked style says what it needs in a new
  `nd` field on its `bar` entry ("a hand to hand weapon in each hand").
- The server sends `recast` (spell id, seconds) when a spell with a recast lands; the spell bar
  darkens that slot and counts the seconds down.
- The stock spell bar shades a hit-point spell by hit points, as the plain bar does.

## Impact

`internal/server/rules` (new rule), `internal/server/world/combat.go`, `pcasting.go`,
`internal/server/protocol` (additive: `bar` entries' `nd`, the `recast` message),
`docs/protocol.md`, `web/quickbar.js`, `web/spells.js`, `web/charwin.js`, `web/net.js`,
`web/ui.js`.
