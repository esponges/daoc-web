# Spec Delta

## ADDED Requirements

### Requirement: The player's effects row
The HUD SHALL show the player's lasting effects as a row of icons, each with its time left and its spell name on hover, chosen by the spell's kind. Beneficial and harmful effects SHALL be told apart by the icon's frame. The row SHALL sit where no other HUD window covers it, and SHALL update the moment an effect starts or ends.

#### Scenario: A buff and a debuff
- **WHEN** the player holds a strength buff and a slow
- **THEN** the row shows both with their time left, the slow framed as harmful, and the slow's icon goes the moment it ends

#### Scenario: Nothing covers it
- **WHEN** the stock interface is drawn with its summary, clock and compass
- **THEN** no part of the effects row lies under another window
