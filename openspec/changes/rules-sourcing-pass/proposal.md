# Proposal: rules-sourcing-pass

## Why

The rule book has 135 rules:

| status | count |
| --- | --- |
| documented (a 1.65 source gives the formula) | 26 |
| reference (understood from a reference server) | 66 |
| provisional (this project's number) | 43 |

Provisional and reference numbers are the main way the game can feel "off" against 1.65 even when every feature exists. Two examples:

- hit points, damage and experience are the project's own;
- archery, endurance, health regeneration and the class tables are provisional.

A pass dedicated to sources, with no new features, moves numbers onto documented footing and records where none can be found.

## What Changes

- **Work through every provisional rule,** most-felt first:
  - combat numbers (melee damage, archery, spell damage variance, resist);
  - then regeneration, experience, falling and resting;
  - then the class tables.
- **For each rule:**
  - search the adopted sources and new 1.65-era ones (Grab Bag archives, Herald patch notes and class pages, Allakhazam, period forum posts with numbers);
  - when a source is found, update the formula if it differs, add the source to the rule's spec, CLAUDE.md and the README, update the worked examples and tests, and set the status to documented;
  - when none is found, write down what was searched in the rule's spec, so the search is not repeated.
- **Reference rules** get the same pass, aiming for documented.
- **Census:** `cmd/parity` adds a "searched, no source" count, so the rules line reads documented, reference, provisional, and how many were searched without result.
- **No new mechanics.**

## Capabilities

### New Capabilities
- (none)

### Modified Capabilities
- `rule-book`: each rule records the sources searched when it stays provisional, and the census counts rules searched without a source.

## Impact

- `openspec/specs/rules-*`, `internal/server/rules/*` and their tests, the registry, `cmd/parity`, README and CLAUDE.md.
- Players may notice changed numbers. Each change is listed in the change's design log.
