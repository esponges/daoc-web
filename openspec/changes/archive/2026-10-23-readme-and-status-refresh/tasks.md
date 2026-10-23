## 1. Archive finished work

- [x] 1.1 Archive `midgard-playable` and `next-midgard-classes-and-spells` with the openspec archive workflow; verify `openspec list` shows neither and `openspec validate --specs` passes.

## 2. Move the reference out of the README

- [x] 2.1 Create `docs/formats.md` from the README's "File formats", "Animation", "NPCs", "Combat" (client side) and converter sections, verbatim; verify nothing was dropped by comparing heading lists before and after.
- [x] 2.2 Fix every `README.md#...` and in-page `](#...)` link to the new location; verify with a grep that no link targets a missing anchor.

## 3. Rewrite the README

- [x] 3.1 Write the new overview: what the game is, run it, multiplayer settings, tests; verify the first screen names the server, classes and client and no "proof of concept" remains (grep).
- [x] 3.2 Add the status table from `docs/parity/parity.json` (one current reading, not-started rows included); verify against `go run ./cmd/parity` output.
- [x] 3.3 Rewrite Known limitations against the code; verify each removed item (casting, creation choices, spells button) by reading the code or by driving the browser client.
- [x] 3.4 Add a test that the README's implemented/total figures equal `parity.json`; verify `go test ./cmd/parity` passes and fails when a figure is changed.
- [x] 3.5 Update CLAUDE.md's "Open now" line and layout table if they changed; verify by reading it.

## 4. Path to level 50

- [x] 4.1 Write `findings.md` in this change ranking what blocks a friends group from levelling 1 to 50 (gear above level 30, trainers and specs, remaining Midgard classes, RvR last), from the item, trainer and class data; verify each ranked item cites its source row or file.
- [x] 4.2 Run `openspec validate readme-and-status-refresh`; verify it passes.

## Workflow follow-up

- Propose `path-to-fifty` from `findings.md` once the target (every Midgard class, or enough for the friends playing) is decided.
