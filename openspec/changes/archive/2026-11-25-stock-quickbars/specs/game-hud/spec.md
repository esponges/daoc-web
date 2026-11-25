## MODIFIED Requirements

### Requirement: Style and quick bar
The HUD SHALL show three quickbars of ten slots, as the install's `default.ini` places Quickbar, Quickbar2 and Quickbar3, drawn from the skin's art, each with ten pages turned by arrows beside it or the mouse wheel and showing its page number. Any slot SHALL hold a style, a spell or an ability, with a style drawn by its weapon line, a spell by its kind and an ability by what it is used with plus a short name. A slot SHALL be used on click or its key: 1 to 0 for the first bar, Shift with them for the second, Alt for the third; 1 SHALL remain the attack key while the first bar's first slot is empty. A slot's tooltip SHALL name it, its opener or cost, and its key. The bars SHALL shade a slot that cannot be used now, show one the character cannot pay for in red, show a recast shade with the seconds left, shade a style held back for the weapon, and ring a song playing or a buff held. Styles, spells and abilities SHALL be placed by dragging them from their windows, moved by dragging between slots and cleared with Shift+Right-click. What the character newly gains SHALL be placed once in the first empty slot of its bar (styles the first, spells the second, abilities the third) and not put back once removed. The layout SHALL be kept per character by the client, and SHALL fall back to plain slots without the converted interface.

#### Scenario: Use a style
- **WHEN** the player clicks a style icon
- **THEN** the style message is sent with the style's id and the icon shows a recast shade until the server allows it again

#### Scenario: Cast from a slot
- **WHEN** the player presses the key of a slot holding a spell
- **THEN** the spell is cast at the current target and the slot shades until the spell can be cast again

#### Scenario: An ability on the third bar
- **WHEN** a Warrior with Guard presses Alt+1 and Guard is in that slot
- **THEN** Guard is used as /guard would, and the server's answer is shown

#### Scenario: Pages
- **WHEN** the player clicks the down arrow beside the second bar
- **THEN** the bar shows its second page, numbered 2, and its keys use that page's slots

#### Scenario: A held-back style
- **WHEN** a style needing a shield sits on a bar and the character holds none
- **THEN** the slot is shaded, and using it shows that it needs a shield
