# player-trade Specification

## Purpose
Two players in one zone exchanging items and money face to face, moved in one step or not at all.

## Requirements

### Requirement: Opening a trade
A player SHALL be able to open a trade with a living player in the same zone within 512 units who is not already trading. Both SHALL see the trade window.

#### Scenario: Trade opened
- **WHEN** Thora targets Bjorn, who stands 200 units away, and uses /trade
- **THEN** both see the trade window with empty offers

#### Scenario: Too far
- **WHEN** the target is 1000 units away
- **THEN** it is refused with "You are too far away to trade."

### Requirement: Offers and accepts
Each side SHALL offer up to 10 items from its bag and an amount of money it holds. Any change to either offer SHALL clear both accepts.

#### Scenario: Changed offer clears accept
- **WHEN** Thora accepts and then Bjorn adds an item
- **THEN** Thora's accept is cleared and both see the new offer

### Requirement: The transfer happens in one step
When both have accepted, the server SHALL move every offered item and the money in one step. If either side lacks bag space or no longer holds what it offered, nothing SHALL move and both SHALL be told why.

#### Scenario: Completed trade
- **WHEN** both accept an offer of a sword for 5 silver
- **THEN** the sword is in Bjorn's bag, the 5 silver in Thora's purse, and the window closes

#### Scenario: No room
- **WHEN** both accept but Bjorn's bag is full
- **THEN** no item or money moves and both read "The trade cannot be completed: not enough room."

### Requirement: Trade ends on interruption
A trade SHALL be cancelled when either side moves out of range, dies, enters combat, logs out or cancels.

#### Scenario: Walking away
- **WHEN** Bjorn walks out of range during a trade
- **THEN** the trade is cancelled for both and nothing moves
