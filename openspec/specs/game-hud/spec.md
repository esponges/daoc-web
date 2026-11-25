# game-hud Specification

## Purpose
TBD - created by archiving change stock-ui. Update Purpose after archive.

## Requirements

### Requirement: Chat window
The HUD SHALL show chat in the stock chat window, send typed lines over the existing chat message, and colour lines by channel. The window's tabs SHALL filter what it shows: Main SHALL show every line, including combat, spells and system lines, and the other tabs SHALL show their own channels. The window SHALL keep at least 200 earlier lines that the player can scroll back through with its scrollbar or the mouse wheel. It SHALL be resizable and movable, and start large enough to show at least twelve lines.

#### Scenario: Say
- **WHEN** a player types a line and presses Enter
- **THEN** the server broadcasts it and every nearby chat window shows it

#### Scenario: Scrolling back
- **WHEN** a fight has pushed earlier lines out of view and the player scrolls up
- **THEN** the earlier lines show again, and a new line does not jump the view back down until the player scrolls to the bottom

#### Scenario: Tabs
- **WHEN** the player selects the Chat tab during a fight
- **THEN** only spoken lines show, and selecting Main again shows the combat lines too

### Requirement: Target and health bars
The HUD SHALL show the player's own health, endurance and mana and the target's name, level colour and health, updated from snapshots.

#### Scenario: Taking damage
- **WHEN** the player is hit
- **THEN** the health bar drops to the server's value

### Requirement: Style and quick bar
The HUD SHALL show three quickbars of ten slots, as the install's `default.ini` places Quickbar, Quickbar2 and Quickbar3, drawn from the skin's art, each with ten pages turned by arrows beside it or the mouse wheel and showing its page number. Any slot SHALL hold a style, a spell or an ability, with a style drawn by its weapon line, a spell by its kind and an ability by what it is used with plus a short name. A slot SHALL be used on click or its key: 1 to 0 for the first bar, Shift with them for the second, Alt for the third; 1 SHALL remain the attack key while the first bar's first slot is empty. A slot's tooltip SHALL name it, its opener or cost, and its key. The bars SHALL shade a slot that cannot be used now, show one the character cannot pay for in red, show a recast shade with the seconds left, shade a style held back for the weapon, and ring a song playing or a buff held. Styles, spells and abilities SHALL be placed by dragging them from their windows, moved by dragging between slots and cleared with Shift+Right-click. What the character newly gains SHALL be placed once in the first empty slot of its bar (styles the first, spells the second, abilities the third) and not put back once removed. Each bar SHALL be movable like a window, by a grip on its page number, keeping its place and returning to its default with /resetui. The layout SHALL be kept per character by the client, and SHALL fall back to plain slots without the converted interface.

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

#### Scenario: Moving a bar
- **WHEN** the player drags the third bar by its page number to the side of the screen
- **THEN** the bar stays there after a reload, and /resetui puts it back

#### Scenario: A held-back style
- **WHEN** a style needing a shield sits on a bar and the character holds none
- **THEN** the slot is shaded, and using it shows that it needs a shield

### Requirement: Inventory and gear
The HUD SHALL show the pack and the equipped items over the stock paper doll, the pack as five pages of eight, with a picture for each item drawn by what it is, move items by dragging, and show the server's refusal reason when a change is refused.

#### Scenario: Equip by drag
- **WHEN** the player drags a hammer onto the weapon slot
- **THEN** the equip message is sent, and on success the slot shows the hammer

#### Scenario: Refused item
- **WHEN** the server refuses an equip
- **THEN** the item stays where it was and the reason is shown

### Requirement: No protocol change
The HUD SHALL work over the existing protocol messages.

#### Scenario: Old and new client
- **WHEN** either UI connects
- **THEN** the server behaves identically

### Requirement: Experience bar
The HUD SHALL show the character's level and progress to the next level in the stock experience window: ten coarse marks for each tenth of the level and a fine bar for progress within the tenth being filled.

#### Scenario: Gaining experience
- **WHEN** the server reports more experience for the character
- **THEN** the fine bar fills and, after each tenth, a coarse mark is added

### Requirement: The player's effects row
The HUD SHALL show the player's lasting effects as a row of icons, each with its time left and its spell name on hover, chosen by the spell's kind. Beneficial and harmful effects SHALL be told apart by the icon's frame. The row SHALL sit where no other HUD window covers it, and SHALL update the moment an effect starts or ends.

#### Scenario: A buff and a debuff
- **WHEN** the player holds a strength buff and a slow
- **THEN** the row shows both with their time left, the slow framed as harmful, and the slow's icon goes the moment it ends

#### Scenario: Nothing covers it
- **WHEN** the stock interface is drawn with its summary, clock and compass
- **THEN** no part of the effects row lies under another window
