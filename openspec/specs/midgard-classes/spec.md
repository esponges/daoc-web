# midgard-classes Specification

## Purpose
TBD - created by archiving change next-midgard-classes-and-spells. Update Purpose after archive.

## Requirements

### Requirement: Five more classes
The server SHALL offer the Berserker, Savage, Healer, Shaman and Runemaster as playable classes, each with its primary, secondary and tertiary stats, hit points per level, weapon skill base, spec point multiplier, base class and allowed races, written in our own words from the rule sources and registered with a status. A character SHALL be creatable only as a class the chosen race may be.

#### Scenario: A race that may not
- **WHEN** a creation request names a Kobold Healer
- **THEN** the server refuses it and the page says which classes the race may take

#### Scenario: Class list
- **WHEN** a player opens character creation with a Troll
- **THEN** the classes offered are Warrior, Berserker, Savage, Shaman and others the Troll may be, and not Healer or Runemaster

### Requirement: Specs and styles by class
Each class SHALL train the spec lines the data gives it, and a melee class SHALL learn the styles the data gives it for its weapon specs. A caster SHALL train its spell lines as specs, and the points spent SHALL set the level of spells it can cast from that line.

#### Scenario: Berserker styles
- **WHEN** a Berserker has trained their hand-to-hand and sword specs
- **THEN** the styles window lists the styles of those lines at their levels

#### Scenario: Spell line spec
- **WHEN** a Healer trains Smite
- **THEN** their book gains the Smite spells up to the trained level

### Requirement: Starting kit and start
A new character of any class SHALL be given the items the data's starting equipment gives its class, and SHALL start at the realm's safe start under the `midgard-playable` start rule.

#### Scenario: A Shaman's kit
- **WHEN** a Shaman is created
- **THEN** they wear and carry the data's items for the class and nothing else

### Requirement: Gender
Character creation SHALL ask for male or female. The server SHALL store it on the character, send it with the character to every client, and clients SHALL draw the model and face for that race and gender. A same-race pair of different genders SHALL look different. An existing character SHALL be male.

#### Scenario: A female Norse Healer
- **WHEN** a female Norse Healer and a male one stand together
- **THEN** both clients draw two different models

#### Scenario: Older characters
- **WHEN** a character made before this change logs in
- **THEN** they are male and nothing else about them changes

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
