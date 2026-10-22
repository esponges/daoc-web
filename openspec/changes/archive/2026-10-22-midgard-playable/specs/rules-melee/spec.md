# Spec Delta

## ADDED Requirements

### Requirement: Styles cast their spells on a hit
A style that lands and has spells in the imported style-spell table SHALL cast each of them on its target, with the row's chance. A row whose class is 0 or the attacker's class SHALL apply. A style that misses or is defended SHALL cast nothing. The spell's effect SHALL follow `rules-spell-effects`, and its hit roll SHALL be skipped. Status: reference.

#### Scenario: Slam stuns
- **WHEN** a Warrior's Slam lands and its style-spell row is a stun at chance 100
- **THEN** the target is stunned for the spell's duration

#### Scenario: A parried style
- **WHEN** the same style is parried
- **THEN** no spell is cast

### Requirement: Weapons cast their proc on a hit
A weapon with a proc spell SHALL cast it on its target on a landed blow, with the item's proc chance, or 10% when the item gives none. The proc SHALL roll its hit as a spell from a caster of the item's level. Status: reference for the 10% default.

#### Scenario: A 10% proc
- **WHEN** a weapon with a damage proc and no proc chance lands a blow and the proc roll is 0.05
- **THEN** the proc spell is cast on the target

### Requirement: Creatures carry their offensive procs
A creature whose spell list holds an offensive-proc spell SHALL have it active from spawn. Each of its landed blows SHALL cast the proc's sub-spell with the proc's chance. Status: reference.

#### Scenario: A proc on a creature's hit
- **WHEN** a creature with a damage-over-time offensive proc lands a blow and the proc roll succeeds
- **THEN** its target receives the damage-over-time effect
