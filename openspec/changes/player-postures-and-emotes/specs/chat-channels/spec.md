## ADDED Requirements

### Requirement: The Emote channel
Emote lines SHALL be sent on their own channel ("emote"). The client SHALL colour and tab them as the stock chat window's Emote filter does, and fall back to the say colour without the converted interface.

#### Scenario: An emote line
- **WHEN** a nearby player waves
- **THEN** the line arrives on the emote channel and shows in the Emote colour
