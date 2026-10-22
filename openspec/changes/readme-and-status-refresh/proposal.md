## Why

The README still introduces the project as a proof of concept on one zone and its
Known limitations contradict the code (it says players cast nothing, there is no
creation wizard, only the Warrior has styles). The game is now a multiplayer
Midgard with six classes, spells and a stock interface, and the README opens with
the converter pipeline, not the game. A reader cannot tell how far the project is
from a live 1.65 server.

## What Changes

- Archive the two finished changes (`midgard-playable`, `next-midgard-classes-and-spells`)
  so the specs the README points at are the archived ones.
- Rewrite the README as a short overview: what the game is, how to run it, a status
  table taken from the parity census, and a current Known limitations list.
- Move the file-format, animation and converter reference out of the README into
  `docs/formats.md`, unchanged in substance, and link it.
- Collapse the three parity snapshots in the README into one current reading.
- Bring CLAUDE.md's "Open now" line in step with the archive.
- Write `findings.md` in this change: a ranked list of what blocks a 1 to 50 run
  for a friends group (gear above level 30, trainers and specs, remaining Midgard
  classes, then RvR), as the input to a follow-up change.

No code changes.

## Capabilities

### New Capabilities
- `project-status`: how the repository states its own progress: the README's
  status section is derived from the parity census and agrees with the code.

### Modified Capabilities

None.

## Impact

- `README.md`, new `docs/formats.md`, `CLAUDE.md`.
- `openspec/changes/archive/` gains two directories; `openspec/specs/` gains the specs they carry.
- The follow-up change for the path to level 50 is out of scope here; this change only ranks it.
