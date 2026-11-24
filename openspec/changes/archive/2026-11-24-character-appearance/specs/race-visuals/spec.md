## ADDED Requirements

### Requirement: Hair is drawn
A humanoid's hair style number SHALL select the hair the install gives that race and gender, drawn on the head and tinted by the hair colour. This SHALL apply to players and to humanoid creatures alike. A helm that covers the head SHALL hide the hair, as in the game.

#### Scenario: A town guard
- **WHEN** a player looks at a Norse guard with no helm
- **THEN** the guard wears hair in the colour the data gives

#### Scenario: Helm hides hair
- **WHEN** a character puts on a full helm
- **THEN** their hair is not drawn
