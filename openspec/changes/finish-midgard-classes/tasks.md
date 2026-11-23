# Tasks

## 1. Sources

- [x] 1.1 Read the Savage spells in the community data (cost column, values, durations) and the Camelot Herald Savage class page. Pin the health-cost meaning and each buff's effect. Add the sources to CLAUDE.md and the README.
- [ ] 1.2 Find sources for Berserk, Charge, Enhanced Evade and QuickCast amounts and timers (Herald class pages, Grab Bag). Record the status of each.
- [ ] 1.3 Search for "Call of a thousand storms" and the style with AttackResultRequirement 3. Implement them or keep the reason.

## 2. Rules

- [x] 2.1 `rules`: Savage buff, endurance heal and Taunt entries, with table tests and registry entries.
- [ ] 2.2 `rules`: Berserk, Charge, Enhanced Evade, QuickCast, Danger Sense and Distraction entries, with table tests and registry entries.

## 3. Server

- [x] 3.1 Add the health cost kind to spells (D1). Refuse when health is short. Add the Savage buff kinds to the buff code, with stat effects on resists, damage, evade, parry and swing speed.
- [x] 3.2 `SavageEnduranceHeal` and `Taunt` (hate on the target's list).
- [ ] 3.3 Berserk, Charge, Enhanced Evade and QuickCast as abilities with effects and reuse timers (D2). Tests per scenario.
- [ ] 3.4 The ground target: request, range and sight check, storage, clearing on zone change, `gt` in own state. Protocol additions in `protocol.go`, `docs/protocol.md` and a protocol test.
- [ ] 3.5 Distraction on the ground target, and Danger Sense on the stealth tick (D4).
- [x] 3.6 Serve the Midgard Rogue trainer like the other base trainers.

## 4. Census

- [ ] 4.1 `cmd/parity`: mark Safe Fall and the new kinds and abilities implemented, and update `gapReasons`. Regenerate `docs/parity`, the README table and `parity_test`.

## 5. Client

- [ ] 5.1 Ground-target key and click on terrain, a marker at the target, and the out-of-range message.
- [ ] 5.2 The new abilities in the Skills window and quickbar, their effect icons, and the health-cost display in the Spells window.
- [ ] 5.3 Check in the browser: play a Savage to level 10 with buffs, a Berserker using Berserk, and a Shadowblade using Distraction. Check the fallback UI.
