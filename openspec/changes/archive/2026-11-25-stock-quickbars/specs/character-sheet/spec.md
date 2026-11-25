## MODIFIED Requirements

### Requirement: Skills and styles list the character's own
The skills window SHALL list the character's specialisations with their levels and the points to spend, and their abilities, including the abilities in `rules-abilities`, with the usable abilities draggable onto a quickbar. The styles window SHALL list the styles the character has, by line, with each style's opening and endurance cost, and SHALL let a style, including one held back for the weapon, be dragged onto a quickbar. Both windows SHALL show their lists in the stock window's own list boxes, scrolling when a list is longer than its box, and SHALL show the points to spend in the window's points label.

#### Scenario: A level 10 Warrior
- **WHEN** a level 10 Warrior with Shields 5 opens SKILLS
- **THEN** the window lists their specs with levels and their abilities, including Guard I

#### Scenario: Styles window to bar
- **WHEN** the player drags a style from the styles window onto a quickbar slot
- **THEN** the slot holds that style and the key for that slot uses it

#### Scenario: Abilities window to bar
- **WHEN** the player drags Intercept from the skills window onto a quickbar slot
- **THEN** the slot holds Intercept and the key for that slot uses it

#### Scenario: A long list
- **WHEN** a character has more abilities than the abilities box can show
- **THEN** the box shows a stock scrollbar and the rest can be scrolled into view
