# Proposal: realm-versus-realm-combat

## Why

RvR is the core of the 1.65 endgame, and nothing of it exists yet. Before battlegrounds, keeps or relics can mean anything, players of different realms must be able to fight, and fighting must be rewarded:

- **Realm points** come from kills.
- **Realm rank** is RR1L0 to the 1.65 cap.
- **Realm abilities** are bought with rank points. The community data has a realm ability table (`classxrealmability_atlas`), which needs filtering to 1.65.
- **RvR death rules** differ from creature deaths.

The stock layout already has the `Realm` training tab (`NewTrainRealmInfo`), `RespecRealm` and `ToggleFloatRealmXP`. The character sheet leaves realm points blank (`web/charwin.js`).

## What Changes

- **Hostility.** Players of different realms may harm each other inside RvR areas: frontier zones and battlegrounds. Outside them they may not. This extends the `mayHarm` predicate from `player-trade-and-social`.
- **Player-versus-player rules** in the rule book:
  - hit location matters against players (the census's "fixed hit location" style feature);
  - armour per slot;
  - crowd-control immunity timers after a mesmerize, stun or root ends;
  - the effect of realm rank.
- **Realm points:**
  - a kill's realm point value comes from the victim's level and realm rank;
  - it is shared among the attackers who damaged the victim, in proportion to their damage, and through groups;
  - the value is reduced for a player killed repeatedly in a short time, under the rule book's anti-farming rule.
- **Realm rank:**
  - RR1L0 to the 1.65 cap, by the published realm rank table;
  - shown on the character sheet, in /who and as a title;
  - each level gives a realm skill point, and each rank gives +1 to all skills (1.65 rule, sourced).
- **Realm abilities:**
  - the 1.65 set per class, bought with realm skill points at a trainer's Realm tab;
  - passive ones change stats or rules (Augmented Strength, Toughness, Mastery of Pain and so on);
  - active ones have reuse timers (Purge, Ignore Pain, Determination and so on);
  - a respec is available.
- **Death in RvR:**
  - no experience loss and no constitution loss;
  - the player is told who killed them;
  - release takes them to the realm's frontier bind point, or the battleground portal keep.

## Capabilities

### New Capabilities
- `rules-realm-points`: realm point value, sharing, the anti-farm reduction, and the realm rank table.
- `realm-abilities`: the 1.65 realm ability set per class, buying them, effects, reuse and respec.
- `rvr-combat`: where players of different realms may fight, player-versus-player defences, crowd-control immunity, and death in RvR.

### Modified Capabilities
- `character-sheet`: realm points, realm rank and title shown.

## Impact

- **Data:** `players.db` adds realm points, rank and bought abilities.
- **Rules:** new rule-book entries for realm points, rank and each realm ability, with tests and the registry.
- **Server:** `world/rvr.go` and `world/realmabilities.go`; combat uses hit location against players.
- **Import:** realm abilities from `classxrealmability_atlas`, filtered to 1.65.
- **Protocol (additive):** realm points and rank on own state, an RP award event, and realm ability lists.
- **Client:** the Realm training tab, realm ability icons and quickbar slots, and the realm-point float text.
