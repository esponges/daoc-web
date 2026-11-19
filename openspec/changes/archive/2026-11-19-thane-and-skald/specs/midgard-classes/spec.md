## ADDED Requirements

### Requirement: Thane and Skald
The server SHALL offer the Thane (class 21) and the Skald (class 24) as playable classes. The Thane SHALL be open to Norseman, Troll and Dwarf, and the Skald to Norseman, Troll, Dwarf and Kobold, as the data's start table gives. Both SHALL have the Viking as their base class. A creation request for any other race SHALL be refused.

#### Scenario: A Kobold Thane
- **WHEN** a creation request names a Kobold Thane
- **THEN** the server refuses it, and the creation page does not offer Thane for a Kobold

#### Scenario: A Kobold Skald
- **WHEN** a player creates a Kobold Skald
- **THEN** the character is made and starts at the realm's safe start with the data's kit for the Skald

#### Scenario: The creation page grows
- **WHEN** a player opens character creation
- **THEN** the classes offered for a Norseman include Thane and Skald

### Requirement: Thane and Skald train their lines
A Thane SHALL be taught by Thane and Viking trainers and SHALL train the spec lines the data gives the class, Stormcalling among them. A Skald SHALL be taught by Skald and Viking trainers and SHALL train its lines, Battlesongs among them. Spells of those lines SHALL enter the book as the spec rises. The Thane and Skald trainer kinds SHALL count as implemented in the census.

#### Scenario: Stormcalling at 5
- **WHEN** a Thane trains Stormcalling to 5
- **THEN** their book holds Thor's Minor Bolt, Thor's Vigor, Thor's Lesser Bolt, Thunder's Bash, Thunder Shout and Toothgrinder's Hoof

#### Scenario: A Skald trainer
- **WHEN** a Skald talks to a Skald trainer
- **THEN** the trainer window lists the Skald's specs; and a Warrior talking to the same trainer is not taught

#### Scenario: The census counts the trainers
- **WHEN** `go run ./cmd/parity` runs on the imported data
- **THEN** the Thane and Skald trainer kinds are listed as implemented
