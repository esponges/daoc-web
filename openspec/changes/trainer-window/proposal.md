# Proposal: trainer-window

## Why

A player reported that there is no way to train a spec at the proper trainer (a Savage at the Savage
trainer raising Hand to Hand with specialisation points). Checking it on 2026-11-24:

- **The rule works.** `rules-specialization` ("Training happens at a trainer of the class") is
  built: `world/services.go` opens a trainer, `Train` raises a spec atomically, and
  `TestTrainingAtAWarriorTrainer` covers it. The imported world has Savage trainers (Ronja in zone
  103, Pirkko and Jaakko in 120, Tuukka 151, Veikko 152, Seaja 154) and Viking trainers that also
  teach a Savage (Skapi in Mularn, 100). The Savage's class specs load: Axe, Hammer, Sword, Hand to
  Hand, Parry, Savagery.
- **The window works once open.** The stock Train window, filled with a Savage's offer, lists the
  six specs; clicking Hand to Hand and pressing TRAIN sends `train` for it (checked in the browser).
- **Nothing tells the player how to get there or what to do.** That is the gap:
  - Right-clicking a trainer opens the Train window silently: no greeting, no word that this is the
    place to spend points. Nothing else in the game says right-click a trainer.
  - Right-clicking the wrong trainer says only "&lt;name&gt; cannot teach you." It does not say whose
    trainer to look for.
  - Gaining specialisation points says how many, never where to spend them.
  - In the window, name, level and cost run together in one column ("Hand to Hand 1   2 pts"),
    and "12 points to spend" overflows its header box.
  - TRAIN with no row chosen, or with a greyed row chosen, does nothing and says nothing.
    Double-clicking a row does nothing.
  - INFO is greyed ("Not built yet."), so there is no way to see what the next level gives (a
    style, an ability, a spell) before spending.
  - Walking away leaves the window open; training from it then fails with "You need to be near a
    trainer of your class."

## What Changes

- A trainer greets the player when right-clicked, in the NPC talk window, and the Train window opens
  beside it. The greeting says how many points are unspent.
- The wrong trainer's refusal names the trainers that teach the player's class.
- The level-up line that grants specialisation points adds where to spend them.
- The Train window shows name and level in the first column and the cost in the second, the points
  in a label that fits, greyed rows with the reason as a tooltip, double-click to train one level,
  a log line when TRAIN cannot act, and INFO listing what the chosen spec's next levels give.
- The Train window closes when the player leaves the trainer's reach.
- The plain fallback panel gets the same greeting, columns, messages and info.

## Capabilities

### New Capabilities
- `trainer-window`: how a player finds out about, opens, reads and uses training.

### Modified Capabilities
- (none; the training rule in `rules-specialization` is unchanged)

## Out of scope

- Choosing a class at level 5 at the base class's trainer: characters are created in their class.
- Respecialisation, and realm abilities (the Train window's lower "Realm Skills" list stays empty;
  `realm-versus-realm-combat` owns it).

## Impact

Server: `world/services.go` (greeting, refusal text, reach check), `world/combat.go` (level-up line),
protocol additive fields on `trainer` (`say`, per-spec `gives`), `docs/protocol.md`.
Client: `web/trainwin.js`, `web/sheet.js` (`trainerView`), `web/ui.js` (greeting, fallback panel,
closing on reach), node tests. No rule-book change.
