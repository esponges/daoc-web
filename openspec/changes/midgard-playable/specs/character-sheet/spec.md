# Spec Delta

## Purpose

Builds the stock attributes, skills and styles windows that the menu bar names, filled from what the server tells a character about themselves, with the plain character sheet as the fallback.

## ADDED Requirements

### Requirement: Character windows open from the menu bar
The ATTRIB., SKILLS and STYLES buttons of the stock menu bar SHALL open stock windows built from their layouts (`StatsAttributes` and `Stats_Spec_Abil`), and a second press SHALL close them. The `C` key SHALL open the attributes window when the stock UI is present.

#### Scenario: Open the attributes window
- **WHEN** the player presses ATTRIB. on the menu bar
- **THEN** the stock attributes window opens with the character's name, level, race, class, eight stats, hit points, endurance, armour factor and weapon skill

#### Scenario: Close it again
- **WHEN** the player presses ATTRIB. again
- **THEN** the window closes

### Requirement: Skills and styles list the character's own
The skills window SHALL list the character's specialisations with their levels and the points to spend, and their abilities, including the abilities in `rules-abilities`. The styles window SHALL list the styles the character has, by line, with each style's opening and endurance cost, and SHALL let a style be dragged to the style bar.

#### Scenario: A level 10 Warrior
- **WHEN** a level 10 Warrior with Shields 5 opens SKILLS
- **THEN** the window lists their specs with levels and their abilities, including Guard I

#### Scenario: Styles window to bar
- **WHEN** the player drags a style from the styles window onto the style bar
- **THEN** the bar holds that style and the key for that slot uses it

### Requirement: Windows show only what the server knows
A character window SHALL show a value only if the server has sent it, and SHALL leave a field empty rather than invent one. The data a window needs SHALL be sent only to the character it describes.

#### Scenario: A field the server does not track
- **WHEN** a layout has a field for something the server does not send, such as realm rank
- **THEN** the field is empty

#### Scenario: Another player's stats
- **WHEN** a player stands beside another
- **THEN** no message sent to them carries the other's stats, specs or abilities

### Requirement: Fallback to the plain sheet
Without converted UI data, the `C` key SHALL open the plain character sheet with the same values, as it does now.

#### Scenario: No converted UI
- **WHEN** `web/data/ui/` is absent and the player presses `C`
- **THEN** the plain character sheet opens and nothing errors
