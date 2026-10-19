# Spec Delta

## Purpose

Defines what dying to a creature costs in classic 1.65: experience lost at death and on release, a gravestone holding part of it that praying recovers, constitution lost on release, and healers who restore it for money.

## ADDED Requirements

### Requirement: Dying to a creature costs experience
When a character of level 6 or more is killed by a creature, they SHALL lose a share of their current level's span: the span multiplied by `p` / 1000, rounded down.
- `p` is 50 − level below level 40, and 10 from level 40.
- For the first death since the character last levelled, `p` is divided by 3.
- For the second, `p` is multiplied by 2/3.
- Each step rounds down.

Experience SHALL never fall below the start of the current level. Status: reference.

#### Scenario: First death at level 10
- **WHEN** a level 10 Warrior with 300,000 experience dies to a creature for the first time since levelling
- **THEN** `p` is 13 and they lose 3,328 experience

#### Scenario: Third death at level 10
- **WHEN** they die a third time at level 10
- **THEN** `p` is 40 and they lose 10,240

#### Scenario: Below level 6
- **WHEN** a level 5 character dies
- **THEN** they lose no experience and no constitution

#### Scenario: At the start of a level
- **WHEN** a character who has just levelled dies
- **THEN** their experience stops at the level's threshold and they keep their level

### Requirement: Releasing costs the same again, into a gravestone
When a dead character releases, the same amount as at death SHALL be lost again, under the same floor. A gravestone holding the amount actually lost SHALL be placed where the character died. A character SHALL have at most one gravestone; a new one SHALL replace the old. The gravestone SHALL be saved with the character so it survives a restart. Status: reference.

#### Scenario: Release after the first death
- **WHEN** the Warrior above releases
- **THEN** they lose 3,328 more, and a gravestone holding 3,328 appears where they fell

#### Scenario: A second grave
- **WHEN** they die and release again before praying
- **THEN** the old gravestone vanishes and the new one holds only the new loss

### Requirement: Praying at your gravestone returns its experience
A living character who targets their own gravestone, stands within 2,000 units of it and asks to pray SHALL, after 5 seconds without moving or fighting, regain all the experience it holds. The gravestone SHALL then disappear. Moving, attacking or being attacked SHALL cancel the prayer. Other players SHALL see the gravestone but SHALL NOT pray at it. Status: reference.

#### Scenario: Pray
- **WHEN** the Warrior walks back, targets their gravestone and prays for 5 seconds
- **THEN** they regain 3,328 experience and the gravestone is gone

#### Scenario: Someone else's grave
- **WHEN** another player targets it and asks to pray
- **THEN** they are told it is not their grave and nothing happens

### Requirement: Releasing costs constitution
Releasing after a creature death at level 6 or more SHALL lower the character's constitution by the death count since levelling: 1 for the first death, 2 for the second, 3 for each later one. Levelling SHALL reset the count. The total lost SHALL be saved, SHALL lower hit points through constitution, and SHALL show on the character sheet. Status: reference.

#### Scenario: Two deaths
- **WHEN** a level 12 Warrior dies and releases twice without levelling
- **THEN** their constitution is 3 lower and their maximum hit points drop accordingly

### Requirement: Healers restore lost constitution for money
Interacting with a healer NPC SHALL tell a character with lost constitution what restoring it costs: the per-point price for their level times the points lost. Accepting SHALL restore as many points as their money pays for, and SHALL take that money. The price per point by level SHALL be the published table: 1, 3, 6, 10, 15, 21, 33, 53, 82, 125, 188 copper for levels 0 to 10, rising to 59,195 at level 50. Status: documented.

#### Scenario: Restore at level 12
- **WHEN** a level 12 Warrior who lost 3 constitution accepts at a healer with 20 silver (2,000 copper)
- **THEN** they pay 1,056 copper and their constitution is back

#### Scenario: Not enough money
- **WHEN** they have only 500 copper
- **THEN** they pay 352 and 1 point is restored

#### Scenario: Nothing to restore
- **WHEN** a character with no lost constitution interacts with a healer
- **THEN** the healer says there is nothing to restore

### Requirement: A released character returns whole
A character who releases SHALL stand at their bind point with full hit points and endurance. Status: reference.

#### Scenario: Back at the bind stone
- **WHEN** a dead Warrior releases
- **THEN** they appear at their bind point at full health
