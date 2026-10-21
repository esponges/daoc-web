## ADDED Requirements

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
