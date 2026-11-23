# Proposal: finish-midgard-classes

## Why

All twelve Midgard classes can be created and trained, but the parity census (`go run ./cmd/parity`) still shows gaps in what their data asks for:

- **Class spell types: 64 of 73.** All nine missing types are the Savage's own:
  - the buffs `SavageCrushResistanceBuff`, `SavageSlashResistanceBuff`, `SavageThrustResistanceBuff`, `SavageDPSBuff`, `SavageEvadeBuff`, `SavageParryBuff` and `SavageCombatSpeedBuff`;
  - `SavageEnduranceHeal`;
  - `Taunt`.

  As things stand, a Savage trains Savagery and gets nothing from it.
- **Abilities: 21 of 30.** Missing are:
  - Berserk (Berserker);
  - Charge (Berserker, Savage);
  - Enhanced Evade (Berserker, Savage);
  - QuickCast (Bonedancer, Runemaster, Spiritmaster);
  - Danger Sense, Distraction, Climb Walls and Safe Fall (Shadowblade);
  - Call of a thousand storms (Thane).
- **Safe Fall is miscounted.** Falling damage arrived with `player-postures-and-emotes`, and `rules.FallDamage` already reduces damage by Stealth spec. The census still lists Safe Fall as "no fall damage for it to reduce".
- **Style features: 13 of 14.** One style opens "after result 3", which no source identifies yet.
- **Ground targets are missing.** Distraction needs them, and later so do ground-targeted area spells and siege engines (`frontiers-and-keeps`).
- **The Midgard Rogue trainer** is one placement the census lists as not implemented.

## What Changes

- **Savage spell kinds.** Each is added with a rule-book entry:
  - the self-buffs pay with health, not power, as the 1.65 Savage did;
  - Taunt adds hate on the target;
  - the endurance heal restores endurance.
- **Berserker and Savage abilities:**
  - Berserk: a timed stance with a damage bonus, lower defence and a reuse timer;
  - Charge: a timed speed boost that combat does not cancel, with a reuse timer;
  - Enhanced Evade: a raised evade chance.
- **QuickCast:** the next spell casts instantly, with a reuse timer, and can be interrupted only by its own rules.
- **Safe Fall:** the census counts it, since the falling rules already apply it. The ability is shown on the character sheet.
- **Ground target:**
  - a player can set a ground target by clicking terrain with the ground-target key (as the stock client does);
  - the target is shown, kept on the server, and limited to a range;
  - Distraction uses it.
- **Danger Sense:** the Shadowblade is warned when a hidden player of an enemy realm comes within range. It is inert until realm-versus-realm play exists, but the rule and message are in place.
- **Climb Walls:** left out, with the reason kept: the 1.65 world has no climbable walls outside keeps. It is revisited by `frontiers-and-keeps`.
- **Call of a thousand storms:** researched. It is implemented if a source names its effect, and marked not classic otherwise.
- **The opening "after result 3":** researched, or left recorded as unidentified.
- **The Midgard Rogue trainer:** served like the other base-class trainers.

## Capabilities

### New Capabilities
- `ground-target`: setting, showing and validating a ground target, and its range.

### Modified Capabilities
- `rules-abilities`: Berserk, Charge, Enhanced Evade, QuickCast, Danger Sense, Distraction and Safe Fall entries.
- `rules-spell-effects`: the Savage spell kinds, Taunt and the endurance heal.
- `midgard-classes`: the Savage is complete, and the Midgard Rogue trainer serves.

## Impact

- **Rules:** `internal/server/rules` gains abilities and spell-effect entries, with table tests and registry entries.
- **Server:**
  - `world/abilities.go`, `world/pcasting.go` and `world/buffs.go` for Savage costs paid in health;
  - `world/catalog_spells.go`;
  - `world/abilities.go` for the timed stances.
- **Census:** `cmd/parity` gap reasons are updated, and the Safe Fall count is fixed.
- **Protocol (additive):** a `groundTarget` request and a `gt` field on the player's own state.
- **Client:** ground-target picking in `viewer.js` and `ui.js`, the marker, and ability buttons in the Skills window.
