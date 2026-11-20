## ADDED Requirements

### Requirement: Resurrection spares the release
A character who is resurrected before releasing SHALL skip the second stage of a death: no second experience loss, no constitution loss and no gravestone. What the death itself took stays lost. Status: documented (the spell's own text in the data says it "eliminates the experience penalty and con loss they would have suffered were they to have /release").

#### Scenario: Raised at level 20
- **WHEN** a level 20 character who lost experience at death is resurrected
- **THEN** their experience is what the death left, with no second loss, no constitution loss and no gravestone

### Requirement: Resurrection sickness
A resurrected character SHALL carry resurrection sickness for 60 seconds, lowering the damage they deal and the strength of their spells by 30 percent; a resurrection that restores all hit points halves the time. It SHALL show in their effects list. Status: reference (the numbers are the data's own spell, the halving is read from DOLSharp).

#### Scenario: Sickness shows
- **WHEN** a character is resurrected
- **THEN** their effects list shows resurrection sickness with its remaining time

### Requirement: Waiting for a resurrection
A dead character SHALL be offered a resurrection and have 15 seconds to accept it; no answer is a decline. They SHALL be able to release at any time after 10 seconds of being dead, and release SHALL be forced after 15 minutes. A pending offer SHALL lapse when they release. Status: documented for the 15 seconds (Camelot Herald patch notes), reference for the other two (DOLSharp).

#### Scenario: Release while an offer is pending
- **WHEN** a dead character has an offer and releases instead
- **THEN** the offer lapses and the release costs what it always does
