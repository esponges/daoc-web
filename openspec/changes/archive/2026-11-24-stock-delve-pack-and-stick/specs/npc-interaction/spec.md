## ADDED Requirements

### Requirement: An NPC turns to the player who interacts with it
When a player interacts with an NPC in reach, the server SHALL turn the NPC to face that player, keep it facing them while it is being talked to, and turn it back to its spawn heading once nobody has interacted with it for 15 seconds.

#### Scenario: Talking to a merchant
- **WHEN** a player right-clicks a merchant standing with its back to them
- **THEN** the merchant turns to face the player and the shop opens

#### Scenario: Left alone
- **WHEN** 15 seconds pass with no interaction
- **THEN** the NPC turns back to its spawn heading
