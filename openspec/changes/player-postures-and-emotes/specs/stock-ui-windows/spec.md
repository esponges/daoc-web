## ADDED Requirements

### Requirement: The Command window
The client SHALL build the stock Command window from its template. SIT, SPRINT, ATTACK, PRAY and RELEASE SHALL send their actions, and FOLLOW, STICK, FACE and BONUSES SHALL be greyed with a tooltip saying why. Without the converted interface, plain buttons SHALL offer sit and sprint.

#### Scenario: Sit from the window
- **WHEN** the player presses SIT in the Command window
- **THEN** they sit, and pressing it again stands them up

#### Scenario: A button not built
- **WHEN** the player hovers FOLLOW
- **THEN** it is greyed and its tooltip says following is not built yet
