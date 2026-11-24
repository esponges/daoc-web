## ADDED Requirements

### Requirement: Appearance choices at creation
Character creation SHALL offer each choice the install provides for the chosen race and gender:
- face;
- hair style;
- hair colour;
- eye colour.

The server SHALL accept only values valid for that race and gender.

#### Scenario: Choosing hair
- **WHEN** a player creates a female Norse character with hair style 3 and colour 2
- **THEN** the character wears that hair in the creation preview and in the world

#### Scenario: Invalid choice
- **WHEN** a creation request names a face the race does not have
- **THEN** creation is refused with the reason

### Requirement: Everyone sees the same appearance
Other players SHALL see a character's chosen face, hair and colours.

#### Scenario: Seen by another
- **WHEN** Bjorn looks at Thora
- **THEN** he sees the hair and face she chose
