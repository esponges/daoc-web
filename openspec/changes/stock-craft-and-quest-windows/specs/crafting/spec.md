## ADDED Requirements

### Requirement: Stock crafting windows
When the converted interface is present, the crafting window SHALL be drawn from the stock Craft and Tradeskills layouts, and fall back to the plain panel when it is absent.

#### Scenario: Fallback
- **WHEN** `web/data/ui/` is absent
- **THEN** the plain crafting panel opens as before
