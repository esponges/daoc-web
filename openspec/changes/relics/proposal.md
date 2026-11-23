# Proposal: relics

## Why

Relics were the top prize of 1.65 RvR. Each realm had two, a Strength relic and a Power relic, each kept in a relic temple guarded by keeps. A realm holding enemy relics gained a damage or casting bonus. `relic.sql` places six relics: two each in regions 1, 100 and 200. The census counts them at zero.

This change depends on `frontiers-and-keeps`: the relic temple gates open according to which keeps a realm holds.

## What Changes

- **Relic temples:**
  - each realm's two temples hold its relics on their pads;
  - the temple gates open to enemies only when the realm has lost the keeps that guard them, by the 1.65 rule.
- **Carrying:**
  - a player picks up a relic;
  - while carrying it, the player can't use speed buffs, can't use teleports and is shown on the war map;
  - dropping it on death leaves it where they fell;
  - it returns to its home temple if left unattended for the rule book's time.
- **Capture:** a relic placed on a pad in the capturing realm's own temple belongs to that realm.
- **Bonuses:**
  - each enemy Strength relic held raises melee damage for the holder realm by the rule book's percentage;
  - each Power relic raises magic damage the same way;
  - the bonuses are shown in the character's bonuses.
- **Announcements** go to all realms when a relic is taken, dropped, returned or captured.

## Capabilities

### New Capabilities
- `relics`: temples, gate rules, carrying, dropping, returning, capture, the realm bonuses and the announcements.

### Modified Capabilities
- (none)

## Impact

- **Import:** relic rows and temple positions.
- **Converters:** relic and temple figures.
- **Server:** new `world/relics.go`; a damage bonus in the melee and spell rules (rule-book entries).
- **Client:** relic models, the carrier effect, the war map's relic markers (the stock `WarmapStrengthRelicSelect` and the others), and the announcements.
