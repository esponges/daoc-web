# Spec Delta

## MODIFIED Requirements

### Requirement: The merchant window uses the stock interface and falls back
The client SHALL show the shop in the stock `Merchant` window. Its list box SHALL list the page's items, each with its picture by kind or weapon line, name and price. Its page buttons and page label SHALL turn the pages. Its quantity field SHALL set how many to buy, and its coin fields SHALL show the selected item's price. Its buy button SHALL buy the selection. Selling SHALL be a click on a pack item while the shop is open. When `web/data/ui/` is absent the client SHALL show the same rows in the built-in panel style.

#### Scenario: Stock UI present
- **WHEN** the converted interface is present and a `shop` message arrives
- **THEN** the stock Merchant window opens with the merchant's name in its title and the stock in its list box with prices

#### Scenario: Buying a stack
- **WHEN** the player selects arrows, raises the quantity to 3 and presses buy
- **THEN** three packs are asked for and the coin fields showed the price of one pack before buying

#### Scenario: Pages
- **WHEN** a merchant sells more items than one page holds
- **THEN** the page label shows the page and the arrow buttons move between pages

#### Scenario: Stock UI absent
- **WHEN** `web/data/ui/` is not there
- **THEN** a plain panel lists the same stock and buying works the same way

#### Scenario: Closing
- **WHEN** the player closes the window or a `shopclose` arrives
- **THEN** the window hides and selling by click stops
