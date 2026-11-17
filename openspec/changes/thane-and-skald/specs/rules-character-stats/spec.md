## ADDED Requirements

### Requirement: The Thane and Skald definitions
The class table SHALL define the Thane (strength, piety, constitution; casts from piety) and the Skald (strength, charisma, constitution; casts from charisma). Each SHALL carry base hit points, weapon skill base and spec multiplier from the sources in the findings: starting at 720, 360, 15 for the Thane and 760, 380, 15 for the Skald. Status: provisional unless a source confirms them.

#### Scenario: Thane definition
- **WHEN** the server looks up class 21
- **THEN** it is the Thane: strength, piety, constitution; a power pool from piety; races Norseman, Troll and Dwarf; base class Viking

#### Scenario: Skald definition
- **WHEN** the server looks up class 24
- **THEN** it is the Skald: strength, charisma, constitution; a power pool from charisma; races Norseman, Troll, Dwarf and Kobold; base class Viking

#### Scenario: Levels raise the class stats
- **WHEN** a Skald reaches level 6
- **THEN** their strength, charisma and constitution rise by the level rule for primary, secondary and tertiary stats
