## MODIFIED Requirements

### Requirement: Direct damage and bolts vary
A creature's direct-damage or bolt spell SHALL do its `Damage` × a variance uniform between 0.75 and 1.25, then the resist, rounded down; a spell of the no-variance type SHALL use a variance of 1. A player's direct-damage or bolt spell from their book SHALL do its `Damage` × (casting stat + 200) / 275, × a roll uniform between a low end and a high end of 1.25, then the resist, then at most three times the spell's `Damage`, rounded down. The low end is 0.25 + (the caster's spec in the spell's line − 1) / the target's level. A spec above the target's level adds 0.005 per level over, and a caster above the target's level adds 0.02 per level of difference, to both ends; the high end is at least 0.25 and the low end between 0 and the high end. A no-variance spell from a player SHALL do its `Damage` scaled by neither. Status: reference (DOLSharp and OpenDAoC-Core; Mythic's Grab Bag answers confirm that a spec at or above the spell's level gives full effect and that damage starts at +25 percent, with no figures).

#### Scenario: Middle of the range
- **WHEN** a creature's 40-damage direct-damage spell rolls a variance of 1.0 against 10% resist
- **THEN** it does 36

#### Scenario: No variance
- **WHEN** a 40-damage no-variance spell lands on a target with no resist
- **THEN** it does exactly 40

#### Scenario: A fully specced caster
- **WHEN** a level 20 caster with 100 in their casting stat and a spec of 20 casts a 40-damage spell at a level 20 target, rolling the middle of the range
- **THEN** the range is 1.2 to 1.25 and the spell does 53

#### Scenario: A low spec
- **WHEN** the same caster has a spec of 5 and rolls the worst result
- **THEN** the low end is 0.45 and the spell does 19

#### Scenario: The cap
- **WHEN** a level 50 caster with 300 in their casting stat and a spec of 50 casts a 10-damage spell at a level 1 target
- **THEN** the range is far above the cap and the spell does 30

### Requirement: Heals restore health
A creature's heal spell SHALL restore `Value` health to its target, up to its maximum. A player's heal from their book SHALL restore a roll between a least and a most, up to the target's maximum: the most is 1.25 × `Value` and the least is `Value` × (0.25 + (the caster's spec in the spell's line − 1) / the spell's level), at most 1.25, and at least 1. A heal whose value is negative SHALL restore that percent of the caster's maximum health. Status: reference (DOLSharp and OpenDAoC-Core; Mythic's Grab Bag answers confirm that heal variance narrows as spec nears the spell's level, with no figures).

#### Scenario: A creature heals itself
- **WHEN** a creature at 100 of 300 health casts a 120-point heal on itself
- **THEN** it has 220

#### Scenario: A player heal at spec
- **WHEN** a caster with a spec of 20 casts a level 20 heal of 120
- **THEN** it restores between 144 and 150

#### Scenario: A player heal below spec
- **WHEN** a caster with a spec of 5 casts a level 20 heal of 120
- **THEN** it restores between 54 and 150

#### Scenario: A percent heal
- **WHEN** a caster with 720 health casts a heal whose value is −40
- **THEN** it restores 288

## ADDED Requirements

### Requirement: A spell rule's status follows its source
A spell rule SHALL NOT be marked above provisional unless its spec names a source for the formula, and SHALL NOT be marked documented unless that source is a published account of 1.65 itself. Where two sources disagree, the spec SHALL say so and name the one followed.

#### Scenario: A reference rule names its source
- **WHEN** a spell rule is marked reference
- **THEN** its spec names the community server or table the formula was understood from, in the project's own words

#### Scenario: No source found
- **WHEN** no source for a formula can be found
- **THEN** the rule stays provisional and `findings.md` records the search

### Requirement: Spell numbers agree with the imported spell table
For each imported spell of an implemented type, the amount the server would deal or restore for a reference caster and target SHALL be derivable from the spell's own `Damage` or `Value`, and a spell whose amount differs from that derivation SHALL be reported.

#### Scenario: A direct damage spell
- **WHEN** the check runs over a 40-damage direct-damage spell at the middle of its range, with no resist
- **THEN** the amount is within the range the rule gives for that spell's base and the spell is not reported

#### Scenario: A heal
- **WHEN** the check runs over a 120-value heal
- **THEN** the amount is within 0.25 to 1.25 times 120, up to the target's maximum

#### Scenario: A mismatch
- **WHEN** a spell's amount falls outside that range
- **THEN** the check names the spell id and both amounts

#### Scenario: No database
- **WHEN** `server/data/community` is absent
- **THEN** the check is skipped and the rest of the tests run
