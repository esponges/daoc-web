## ADDED Requirements

### Requirement: Every offered class levels
A character of any class the server offers SHALL gain experience, levels, hit points, spec points and, for casters, power by that class's table, and SHALL train specs from trainers who teach their class or its base class.

#### Scenario: A Berserker levels
- **WHEN** a level 1 Berserker gains enough experience for level 2
- **THEN** they gain the hit points and spec points their class table gives

#### Scenario: A Healer trains
- **WHEN** a Healer visits a Healer trainer
- **THEN** the trainer window lists the Healer's spec lines, and trainers of other classes' bases are refused
