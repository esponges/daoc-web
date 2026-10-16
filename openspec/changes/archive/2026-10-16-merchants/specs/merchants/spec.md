## ADDED Requirements

### Requirement: A merchant offers the stock its list names
A merchant's stock SHALL be the items of the imported list its spawn names, in the list's page and slot order, each with the price the item table gives. An item the item table does not contain SHALL be left out of the stock and counted in the server's log.

#### Scenario: Stock is a join
- **WHEN** a merchant's spawn names a list with twelve rows, two of which name unknown items
- **THEN** the `shop` message carries ten items in page and slot order and the log counts two skipped

#### Scenario: A merchant with no list
- **WHEN** a merchant spawn names no list, or a list with no rows
- **THEN** interacting opens an empty shop (nothing to buy; selling still works)

### Requirement: Buying is decided by the server
On `buy` the server SHALL check that the player has an open merchant in reach, that the item is in that merchant's stock, that the player has enough money for the price, and that the pack has a free slot. When all hold it SHALL take the money and add the item in one saved step and send the updated `inv`; otherwise it SHALL change nothing and answer with an `err`.

#### Scenario: A purchase
- **WHEN** a player with enough money buys an item from an open merchant
- **THEN** the price is taken, the item appears in the first free pack slot, `inv` is resent, and the character is saved with both changes

#### Scenario: Too poor
- **WHEN** the player's money is below the price
- **THEN** nothing changes and the error says the player cannot afford it

#### Scenario: Full pack
- **WHEN** the pack has no free slot
- **THEN** no money is taken and the error says the pack is full

#### Scenario: Not in stock
- **WHEN** the request names an item the merchant does not sell, or no merchant is open
- **THEN** the request is refused and nothing changes

#### Scenario: Walked away
- **WHEN** the player is out of interact range when `buy` arrives
- **THEN** the request is refused and the server sends `shopclose`

### Requirement: Selling is decided by the server
On `sell` the server SHALL accept only a pack slot that holds an item, only with a merchant open and in reach, and SHALL pay a fixed fraction of the item's table price (rounded down, at least one copper when the price is positive). Worn items and items with a price of zero SHALL NOT be sold. The item removal and the payment SHALL be saved together.

#### Scenario: A sale
- **WHEN** the player sells a pack item with a positive price
- **THEN** the item is removed, money rises by the sale price, `inv` is resent, and both are saved together

#### Scenario: Worn or worthless
- **WHEN** the slot is a worn slot, empty, or holds an item priced at zero
- **THEN** nothing changes and the error says why

### Requirement: The shop closes when it should
The server SHALL close a player's open shop, with `shopclose`, when the player moves out of interact range, dies, changes zone or disconnects, or when the merchant dies. Opening a second shop SHALL replace the first.

#### Scenario: Walking off
- **WHEN** the player moves beyond interact range of the merchant whose shop is open
- **THEN** the server sends `shopclose` and later `buy` and `sell` are refused until the player interacts again

### Requirement: The merchant window uses the stock interface and falls back
The client SHALL show the shop in a window drawn from the stock frame, with a row per item (picture by kind or weapon line, name, detail and price), a click to buy, and the player's money. Selling SHALL be a click on a pack item while the shop is open. When `web/data/ui/` is absent the client SHALL show the same rows in the built-in panel style.

#### Scenario: Stock UI present
- **WHEN** the converted interface is present and a `shop` message arrives
- **THEN** the merchant window opens over the stock frame listing the stock with prices

#### Scenario: Stock UI absent
- **WHEN** `web/data/ui/` is not there
- **THEN** a plain panel lists the same stock and buying works the same way

#### Scenario: Closing
- **WHEN** the player closes the window or a `shopclose` arrives
- **THEN** the window hides and selling by click stops
