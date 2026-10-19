# Spec Delta

## Purpose

Keeps the game's rules this project's own and checkable: every mechanic the server implements is written down in our words with its sources, and a census measures how much of a classic 1.65 Midgard the server covers.

## ADDED Requirements

### Requirement: Every implemented mechanic has a rule spec
Each game mechanic the server implements SHALL be described by a capability spec named `rules-<mechanic>`: what it decides, its inputs, the formula in this project's words, its caps, and at least one scenario with concrete numbers. Server code that decides a game outcome SHALL implement a rule from such a spec.

#### Scenario: A new mechanic
- **WHEN** a change adds a game mechanic to the server, such as a new kind of defence
- **THEN** the same change adds or modifies a `rules-*` spec that states it, and registers the rule so the census counts it

#### Scenario: Reading a rule
- **WHEN** someone wants to know how the server computes a Warrior's hit points
- **THEN** `rules-character-stats` states the formula and a worked example, without reading the code

### Requirement: Each rule states its source and status
Each rule SHALL name its sources and SHALL carry one status:
- *documented*, when a public source states the rule;
- *reference*, when it matches the classic reference server's behaviour and no public source gives the number;
- *provisional*, when neither does and the value is this project's choice.

CLAUDE.md and the README SHALL list every source adopted.

#### Scenario: A documented rule
- **WHEN** the parry rule cites a public mechanics page for "5% plus 0.5% per spec point"
- **THEN** it is marked documented and that page is listed in CLAUDE.md and the README

#### Scenario: A provisional rule
- **WHEN** no source gives a health regeneration tick and the project picks one
- **THEN** the rule is marked provisional, says so in its text, and the census counts it as provisional

### Requirement: No rule is copied from a GPL server
Rules SHALL be written from their descriptions in this project's words and code. Source code from a GPL emulator SHALL NOT be copied into the repository, its specs or its tests. Reading it to understand a mechanic is allowed and SHALL be named as a reference source.

#### Scenario: Source review
- **WHEN** the rules package and its tests are reviewed
- **THEN** they contain no code taken from a GPL project, and each test names the rule scenario it checks

### Requirement: Rule scenarios are test vectors
Every scenario in a `rules-*` spec that gives numbers SHALL have a matching table test in the server. That test SHALL compute the scenario's inputs with the server's rules and SHALL expect the scenario's outputs.

#### Scenario: A vector drifts
- **WHEN** someone changes the hit point formula and the level 50 Troll Warrior scenario no longer holds
- **THEN** `go test` fails and names the scenario

### Requirement: The parity census measures Midgard against the data
A command SHALL read the community database dump and the server's registry of implemented rules. For Midgard (regions 100, 101 and 151) it SHALL report each of these, with the number of placements or uses that need it:
- the spell types creatures and classes use;
- the style features used;
- the abilities granted;
- the creature behaviours by class type;
- the NPC services present.

For each item it SHALL say whether the server implements it. It SHALL also report the number of rules by status.

#### Scenario: Census run
- **WHEN** the census runs against a checked-out dump
- **THEN** it writes `docs/parity/parity.json` and `docs/parity/README.md`, which list for example that 15 creature spell types are used by 1,763 placements and that none is implemented

#### Scenario: No dump
- **WHEN** the dump is absent
- **THEN** the command says where it looked and exits with an error, and the census test skips

### Requirement: The census output is safe to commit and stable
The census SHALL write only names, counts and statuses: no row content, text, coordinates or paths outside the repository. Two runs on the same dump and code SHALL produce the same bytes.

#### Scenario: Re-run
- **WHEN** the census runs twice with nothing changed
- **THEN** both runs write identical files

#### Scenario: Review of the output
- **WHEN** someone reads `parity.json`
- **THEN** it holds type and class names with counts and statuses, and no spell, creature or quest text

### Requirement: Parity does not go backwards silently
A test SHALL compare a fresh census with the committed baseline. It SHALL fail when an item that was implemented is no longer implemented, or when a rule's status moves away from documented.

#### Scenario: A rule dropped
- **WHEN** a change removes the code for the Warrior's parry rule
- **THEN** the census test fails and names the item that went backwards
