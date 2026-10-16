## ADDED Requirements

### Requirement: Interaction and trade are additive messages
The protocol SHALL add the requests `interact`, `buy` and `sell` and the server messages `shop` and `shopclose`, and MAY add an interaction kind `ik` to NPC entities. Existing messages and fields SHALL NOT change meaning, and a client that does not know the new ones SHALL keep working.

#### Scenario: An older client
- **WHEN** a client that predates this change joins
- **THEN** it ignores `ik`, never sends `interact`, and plays as before

#### Scenario: Round trip
- **WHEN** each new message is encoded and decoded in the protocol tests
- **THEN** it survives unchanged, and `docs/protocol.md` documents every field
