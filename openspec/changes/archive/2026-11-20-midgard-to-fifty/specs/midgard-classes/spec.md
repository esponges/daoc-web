## ADDED Requirements

### Requirement: The last four Midgard classes
Spiritmaster, Bonedancer, Hunter and Shadowblade SHALL be offered at creation for exactly the races the data's start table allows each of them, with male or female. Their trainers in the world SHALL teach them, as for the other classes. With them, all twelve of Midgard's 1.65 classes SHALL be playable.

#### Scenario: Creation page
- **WHEN** the creation page lists Midgard's classes
- **THEN** it offers all twelve, and each of the four new classes only for the races the start table lists for it

#### Scenario: A trainer teaches
- **WHEN** a Bonedancer talks to a Bonedancer trainer with unspent points
- **THEN** the trainer offers the Bonedancer's specs

### Requirement: Each new class has its specs and abilities
Each new class SHALL train the specs the data's class table gives it, and its spell lines and abilities SHALL come from the imported data:
- Spiritmaster: Darkness, Suppression, Summoning;
- Bonedancer: Darkness, Suppression, Bone Army;
- Hunter: Beastcraft, Composite Bow, Spear, Sword, Stealth;
- Shadowblade: Critical Strike, Envenom, Left Axe, Stealth, Axe, Sword.

Every spell and ability these grant SHALL either work, or be counted as not implemented in the census.

#### Scenario: Spellbook of a Spiritmaster
- **WHEN** a level 10 Spiritmaster with Summoning 5 opens their spells
- **THEN** the window lists the spells their lines give at those levels, and each can be cast

#### Scenario: Census
- **WHEN** `go run ./cmd/parity` runs after this change
- **THEN** no spell type or ability used only by these four classes is left unimplemented without a reason
