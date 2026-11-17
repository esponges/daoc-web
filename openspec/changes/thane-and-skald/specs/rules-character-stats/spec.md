## ADDED Requirements

### Requirement: The Thane and Skald definitions
The class table SHALL define the Thane (strength, piety, constitution; casts from piety; 720 hit points, weapon skill base 360, multiplier 20) and the Skald (charisma, strength, constitution; casts from charisma; 760 hit points, weapon skill base 380, multiplier 15). Status: provisional (one source, no 1.65 account).

#### Scenario: Thane definition
- **WHEN** the server looks up class 21
- **THEN** it is the Thane with the numbers above, a power pool from piety, races Norseman, Troll and Dwarf, and base class Viking

#### Scenario: Skald definition
- **WHEN** the server looks up class 24
- **THEN** it is the Skald with the numbers above, a power pool from charisma, races Norseman, Troll, Dwarf and Kobold, and base class Viking

#### Scenario: Levels raise the class stats
- **WHEN** a Skald reaches level 6
- **THEN** their charisma, strength and constitution rise by the level rule for primary, secondary and tertiary stats
