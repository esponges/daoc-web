## ADDED Requirements

### Requirement: Searched sources are recorded
A rule that stays provisional or reference after a search SHALL list in its spec the sources searched and the date. The registry SHALL mark it as searched.

#### Scenario: No source found
- **WHEN** a search for the 1.65 endurance regeneration rate finds nothing
- **THEN** the `rules-character-stats` spec lists the sources tried and the date, and the registry marks the rule searched

### Requirement: Census counts searched rules
The parity census SHALL show, besides documented, reference and provisional, how many rules were searched without a source.

#### Scenario: Census line
- **WHEN** `cmd/parity` runs after the pass
- **THEN** the rules line reads "rules: N documented, N reference, N provisional (N searched, no source)"

### Requirement: A changed number is logged
When a sourced number replaces a provisional one, the rule's worked examples and tests SHALL change with it, and the change's design log SHALL name the old and new values.

#### Scenario: A changed formula
- **WHEN** a source gives a different archery damage formula than the provisional one
- **THEN** the archery spec, its table tests and the design log all show the new formula and the old one it replaced
