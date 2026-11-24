# Design: trainer-window

## Context

Training is server-decided and already correct (`rules-specialization`). Everything here is how
the player learns about it and uses it. Nothing in this change alters a cost, a cap or who may
train; it adds words and makes the stock Train window's own controls do what they look like they do.

## Decisions

### 1. The greeting rides on the `trainer` message

`trainer` gains an optional `say` string: the trainer's greeting in our words, for example
"Welcome, Hrafn. You have 12 specialization points to spend. Choose a specialization and press
TRAIN." With no points: "...You have no specialization points to spend; you gain more as you
level." The client shows it in the NPC talk window (`Interact`, already built for healers, with no
accept button) and opens the Train window next to it. Without the converted interface, the plain
trainer panel shows it as its first line.

Re-sends of `trainer` after a train carry no `say`, so the talk window is not re-opened each level.

The exact 1.65 greeting text is not sourced; ours is our own and says what to do. Status:
provisional.

### 2. The refusal names the trainers

"Ronja cannot teach you. Seek a Savage or Viking trainer." The names come from the class table:
the class and its base (`Class.Name`, `Class.Base`), the same pair `TaughtBy` accepts, so the text
cannot disagree with the rule. Classes without a base name one trainer.

### 3. The level-up line says where

When a level or half level grants points, the line already sent ("You gain 3 specialization
points.") adds " Visit a Savage trainer to spend them." Only the class's own trainer is named, to
keep the line short.

### 4. The Train window's columns and label

The stock template's list (`1002`, adapter `train_skills`) has `Columns` 2. Rows carry the name and
level in the first column ("Hand to Hand 4") and the cost in the second ("5"); `uictl` lists already
draw a second column where a template asks for it — if not, this change adds it there, not in
`trainwin.js`. The header label over the list (the template's `kwijiboe` field) shows "12 points",
which fits its 92-pixel box at arial11.

Greyed rows keep their tooltip; the reason is made specific: "Needs 5 points; you have 3." or
"Hand to Hand cannot go above your level (4)."

### 5. TRAIN, double-click and messages

- TRAIN with nothing chosen: log "Choose a specialization to train first."
- TRAIN on a greyed row: log the row's reason.
- Double-click on a row trains it one level, the same as select and TRAIN.
- The server stays the judge; these messages only save a round trip. Its own refusals still reach
  the log.

### 6. INFO: what the next levels give

`trainer` spec rows gain an optional `gives` list: up to five unlocks above the current level, each
`{lv, nm, k}` where `k` is `style`, `ability` or `spell`. The server has these already (styles by
spec and level, `spec_ability`, the spell lines). INFO (enabled now) writes them to the log under
the spec's name, e.g. "Hand to Hand: 4 Savage Rage style, 6 Mighty Fists style, ...". A spec with
nothing ahead says so. The row tooltip also names the next unlock.

The log rather than a new window: the template has no description area, and the lower list
belongs to realm skills.

### 7. Leaving reach closes the window

The server tracks the open trainer as it tracks an open shop and sends `trainclose` when the player
walks out of reach, dies or the trainer is gone (checked each tick, like `checkShop`). The client
closes the Train window and the greeting it opened. The reach check in `Train` is unchanged.

## Risks

- The `gives` lookup must use the same data the character sheet uses, or INFO and the Styles window
  will disagree. Read it through the catalog functions the sheet already calls.
- Older clients ignore the new fields (protocol additive).
