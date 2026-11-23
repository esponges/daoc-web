# Design: hibernia-realm

## Context

`albion-realm` turns realm into a parameter (its D1). This change adds Hibernia's rows and the class mechanics Hibernia brings.

## Decisions

**D1. Turrets are pets that cannot move.**
- They reuse the pet brain with a zero move speed, a placement at the ground target, and a per-caster limit.
- Each turret's spell comes from its summon spell's data.

**D2. Bomber pets** are pets that run to the target and cast one spell on arrival, then leave.

**D3. Share before adding.**
- Songs, dual wield, charm and stealth reuse the Midgard code paths.
- The census for Hibernia shows only the kinds that are actually new.

**D4. Look at every zone,** as with Albion.

## Non-Goals

- RvR, which belongs to later changes.
