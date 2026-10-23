## ADDED Requirements

### Requirement: README describes the project as it is
The README SHALL open with what the game is today and how to run it, and SHALL NOT describe the project as a proof of concept.

#### Scenario: Opening reflects the game
- **WHEN** a reader opens the README
- **THEN** the first section names the multiplayer Midgard server, the playable classes and the browser client, before any converter detail

#### Scenario: Converter reference is linked, not inlined
- **WHEN** a reader looks for a file format or animation write-up
- **THEN** the README links to `docs/formats.md`, which holds it

### Requirement: Status comes from the parity census
The README SHALL carry one status table whose numbers equal those in `docs/parity/parity.json`.

#### Scenario: Numbers agree
- **WHEN** `go run ./cmd/parity` is run and the README table is compared with its output
- **THEN** every implemented/total figure in the table matches

#### Scenario: One reading, no history
- **WHEN** the README is read
- **THEN** it shows the current reading only, not earlier baselines

### Requirement: Known limitations are true
Each entry in Known limitations SHALL hold for the current code.

#### Scenario: Removed limitations
- **WHEN** a limitation names a feature the code now has (player casting, class and gender choice at creation, the spells button)
- **THEN** that entry is absent or corrected

### Requirement: Remaining distance to live 1.65 is stated
The README SHALL name what is not implemented that a live 1.65 server has, including quests, keeps, relics, battlegrounds, trainers, pets, and the Midgard classes not yet playable.

#### Scenario: Gaps listed
- **WHEN** the status section is read
- **THEN** each parity-census row at zero implemented appears as not started
