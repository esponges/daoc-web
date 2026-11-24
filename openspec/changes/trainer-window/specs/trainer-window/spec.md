## ADDED Requirements

### Requirement: A trainer greets the player and opens training
When a player interacts with a trainer that teaches their class or its base class, the server SHALL send the trainer's offer with a greeting that names the player and their unspent specialization points and says how to train. The client SHALL show the greeting in the NPC talk window and open the Train window. Without the converted interface, the plain trainer panel SHALL show the greeting as its first line. An offer re-sent after training SHALL carry no greeting. Status: provisional.

#### Scenario: A Savage right-clicks a Savage trainer
- **WHEN** a level 10 Savage with 12 unspent points right-clicks Ronja, Savage Trainer, in reach
- **THEN** the talk window shows a greeting naming the Savage and saying they have 12 specialization points to spend, and the Train window opens listing Axe, Hammer, Sword, Hand to Hand, Parry and Savagery

#### Scenario: No points to spend
- **WHEN** a character with no unspent points right-clicks their trainer
- **THEN** the greeting says they have no points to spend and that they gain more as they level, and the Train window still opens

#### Scenario: Training does not repeat the greeting
- **WHEN** the player trains Hand to Hand from an open Train window
- **THEN** the Train window refreshes and the talk window is not opened again

### Requirement: The wrong trainer says which trainers to seek
When a player interacts with a trainer that does not teach them, the refusal SHALL name the trainers that do: their class, and their class's base class when it has one. Status: provisional.

#### Scenario: A Savage at the Healer trainer
- **WHEN** a Savage right-clicks a Healer trainer in reach
- **THEN** the log says that trainer cannot teach them and to seek a Savage or Viking trainer, and no Train window opens

### Requirement: Gaining points says where to spend them
When a level or half level grants specialization points, the line announcing them SHALL also say to visit a trainer of the character's class to spend them. Status: provisional.

#### Scenario: A Savage reaches level 6
- **WHEN** a Savage reaches level 6 and gains specialization points
- **THEN** the log line names the points gained and says to visit a Savage trainer to spend them

### Requirement: The Train window reads at a glance
The Train window SHALL list each trainable spec with its name and current level in the first column and the cost of its next level in the second, show the unspent points in the header label within its box, draw specs that cannot be trained now in grey, and give each row a tooltip that says what the next level costs and, for a grey row, why it cannot be trained. Status: provisional.

#### Scenario: Columns and points
- **WHEN** a Savage with 12 points and Hand to Hand at 4 opens the Train window
- **THEN** the Hand to Hand row reads "Hand to Hand 4" with "5" in the second column, and the header reads "12 points" without overflowing its box

#### Scenario: A spec at the character's level
- **WHEN** a level 4 Savage has Hand to Hand at 4
- **THEN** the row is grey and its tooltip says it cannot go above the character's level

#### Scenario: Not enough points
- **WHEN** the next level of Savagery costs 5 and the character has 3 points
- **THEN** the row is grey and its tooltip says it needs 5 points and the character has 3

### Requirement: Training from the window always answers
Pressing TRAIN with a trainable row chosen, or double-clicking a trainable row, SHALL ask the server to train that spec one level. Pressing TRAIN with no row chosen SHALL log that a specialization must be chosen first; pressing it on a grey row SHALL log that row's reason. The server's own refusals SHALL still reach the log. Status: provisional.

#### Scenario: Double-click
- **WHEN** the player double-clicks the Hand to Hand row and can afford it
- **THEN** a `train` request for Hand to Hand, one level, is sent and the window refreshes with Hand to Hand one level higher and the points reduced

#### Scenario: TRAIN with nothing chosen
- **WHEN** the player presses TRAIN without choosing a row
- **THEN** the log says to choose a specialization first and nothing is sent

### Requirement: INFO shows what a spec's next levels give
The trainer's offer SHALL list, for each spec, up to five styles, abilities or spells it grants above its current level, each with the level that grants it, taken from the same data the character sheet uses. The INFO button SHALL write the chosen spec's list to the log, or say that nothing more is granted. Each row's tooltip SHALL name the next unlock. Status: provisional.

#### Scenario: INFO on Hand to Hand
- **WHEN** the player chooses Hand to Hand and presses INFO
- **THEN** the log lists the Hand to Hand styles above the current level with the level each is gained at, the same names and levels the Styles window shows once trained

#### Scenario: INFO with nothing chosen
- **WHEN** the player presses INFO with no row chosen
- **THEN** the log says to choose a specialization first

### Requirement: Leaving the trainer closes the window
The server SHALL tell the client when the player moves out of the trainer's interact range, dies, or the trainer is no longer present, and the client SHALL then close the Train window and the talk window it opened. Status: provisional.

#### Scenario: Walking away
- **WHEN** the player walks beyond interact range of the trainer with the Train window open
- **THEN** the Train window and the greeting close, and no `train` request can be sent from them
