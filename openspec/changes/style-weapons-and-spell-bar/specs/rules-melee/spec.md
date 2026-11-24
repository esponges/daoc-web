## ADDED Requirements

### Requirement: A style needs its weapons
A style SHALL be usable only with the weapons its requirement names. A requirement that is an object type SHALL need a main-hand weapon of that type, and a shield requirement a shield in the left hand. The requirement "two weapons" SHALL need a weapon in the main hand and a weapon, not a shield, in the left hand; for a Hand to Hand style both SHALL be hand to hand weapons, and for a Left Axe style the left SHALL be a left axe. The requirement "any weapon" SHALL need any weapon in the main hand. Status: reference.

#### Scenario: Claws in both hands
- **WHEN** a Savage holds a hand to hand weapon in each hand
- **THEN** a Hand to Hand style is usable

#### Scenario: One claw
- **WHEN** a Savage holds a hand to hand weapon in the main hand and nothing in the left
- **THEN** a Hand to Hand style is not usable, and it needs a hand to hand weapon in each hand

#### Scenario: An axe and a shield
- **WHEN** a Savage holds an axe and a shield
- **THEN** a Hand to Hand style is not usable

#### Scenario: A left axe
- **WHEN** a Berserker holds a sword and a left axe
- **THEN** a Left Axe style is usable

#### Scenario: Any weapon
- **WHEN** a Shadowblade holds a sword and nothing in the left hand
- **THEN** a Critical Strike style is usable, and with no weapon it is not
