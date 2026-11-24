# Proposal: rules-sourcing-reference

## Why

`rules-sourcing-pass` searched the 44 provisional rules that a public source might cover and found one change. The 66 reference rules (understood from a reference server, not from a period source) were not searched, and the Camelot Herald and fandom class pages failed to load. The research found these still worth reading in full: the Herald Savage, Berserker, Ranger and Shadowblade pages, and the 1.65 and 1.65A to I patch notes.

## What Changes

- Read the full 1.65 and 1.65A to I patch notes and the class pages (try the web archive when the live pages fail); search a period source for each reference rule, most-felt first.
- Where a source is found: update the formula if it differs, add the source to the rule's spec, CLAUDE.md and the README, and set the status to documented.
- Where none is found: add a Searched note.
- Revisit the findings the first pass did not adopt (power regeneration, charm level cap, haste floor, Critical Shot, group challenge) with a second source each.

## Capabilities

### Modified Capabilities
- `rule-book`: more rules documented.

## Impact

`internal/server/rules/*`, their specs, `sourcing.go`, the census.
