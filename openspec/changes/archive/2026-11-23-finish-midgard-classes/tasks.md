# Tasks

## 1. Sources

- [x] 1.1 Read the Savage spells in the community data (cost column, values, durations) and the Camelot Herald Savage class page. Pin the health-cost meaning and each buff's effect. Add the sources to CLAUDE.md and the README.
- [x] 1.2 Find sources for Berserk, Charge, Enhanced Evade and QuickCast amounts and timers (Herald class pages, Grab Bag). Record the status of each. (Searched 2026-11-23: Charge existed as a 1.53 realm ability, QuickCast was on the bar in 1.65, Berserk is described only by a later developer post; none gives amounts, so Berserk, Charge and QuickCast are provisional. Nothing at all on Enhanced Evade: left out, see 2.2.)
- [x] 1.3 Search for "Call of a thousand storms" and the style with AttackResultRequirement 3. Implement them or keep the reason. (Searched 2026-11-23, nothing found. Call of a thousand storms is granted at level 51, above this world's cap; the style is Aurora Borealis. Both stay census gaps with those reasons.)

## 2. Rules

- [x] 2.1 `rules`: Savage buff, endurance heal and Taunt entries, with table tests and registry entries.
- [x] 2.2 `rules`: Berserk, Charge, QuickCast, Danger Sense, Distraction and the ground target, with table tests and registry entries. Enhanced Evade has no entry: no source says what it did (spec: "Enhanced Evade is left out, with its reason").

## 3. Server

- [x] 3.1 Add the health cost kind to spells (D1). Refuse when health is short. Add the Savage buff kinds to the buff code, with stat effects on resists, damage, evade, parry and swing speed.
- [x] 3.2 `SavageEnduranceHeal` and `Taunt` (hate on the target's list).
- [x] 3.3 Berserk, Charge and QuickCast as abilities with effects and reuse timers (D2). Tests per scenario. (Enhanced Evade: see 2.2.)
- [x] 3.4 The ground target: request, range and sight check, storage, clearing on zone change, `gt` in own state. Protocol additions in `protocol.go`, `docs/protocol.md` and a protocol test.
- [x] 3.5 Distraction on the ground target, and Danger Sense on the stealth tick (D4).
- [x] 3.6 Serve the Midgard Rogue trainer like the other base trainers.

## 4. Census

- [x] 4.1 `cmd/parity`: mark Safe Fall and the new kinds and abilities implemented, and update `gapReasons`. Regenerate `docs/parity`, the README table and `parity_test`.

## 5. Client

- [x] 5.1 Ground-target key (T, then click the ground; shift-T clears) and a marker at the target. The server's out-of-range and out-of-sight refusals arrive as errors.
- [x] 5.2 The new abilities in the ability bar with their effect timers, and the health-cost display in the Spells window and the spell bar (both the stock window and the plain panel).
- [x] 5.3 Checked in the browser on a scratch server and database: a Savage trained Savagery and cast a parry buff (health fell by the cost, the effect and its timer showed); the ground target marker was set by clicking; a Berserker used Berserk (log line, effect timer, "Berserk 4" button). The plain-panel fallback was checked with `web/data/ui` absent from a scratch copy of `web`. Distraction and Danger Sense are covered by server tests but were not played in the browser.
