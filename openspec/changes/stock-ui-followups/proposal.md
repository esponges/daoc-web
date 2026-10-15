## Why

`stock-ui` and `stock-ui-rest` put the game's own windows on screen, but they left things out: the menu bar's buttons are greyed because the windows behind them were never built, two flows (a real kill's loot, a real death's release) were exercised only with sample data, the pictures on styles and items are generic, the compass orientation and the clock are unchecked or stand-ins, and a new character walking out of the Myrkwood spawn meets level-26 creatures that kill it in two blows. This change collects those.

## What Changes

- **Spike first:** for each item, record what the server already sends (it sends health, endurance, level and experience, but no attributes, skills or group), what stock layouts exist (`StatsAttributes`, `Stats_Spec_Abil`, `StatsGroup`, `new_group`, `MiniGroup` do), and whether the dangerous-spawn problem is in the imported data or the server's rules.
- **Character windows:** build the attributes, styles, skills and spells windows the menu bar names, over their layouts and the classic character sheet art, showing what the server can honestly tell them; add to the protocol only what a window needs.
- **Group:** decide from the spike whether a group window is in scope or stays disabled until the server has groups.
- **Starting-zone safety:** find why creatures far above a new character sit beside the spawn and aggro it, and fix it in the data's use or the server's rules, not by editing the data.
- **Proof:** see a real kill and its loot through the stock loot window, and a real death through the Release window with an early release, with screenshots in the change.
- **Orientation and time:** check the compass's north against a known landmark; give the clock a game time only if a day length can be established from the install, else keep local time.
- **Icons:** look for a real source of per-style and per-item pictures; if none is found, keep the kind-based ones and say so.
- **Left out on purpose:** chat tabs Broad, Guild and Group until the server has those channels; journal, map and warmap stay greyed with a tooltip saying why.
- Converted art stays gitignored; the built-in UI stays as the fallback.

## Capabilities

### New Capabilities
- `character-sheet`: the attributes, styles, skills and spells windows and what the server tells them.
- `starting-zone-safety`: how creatures of a level far above a new character behave near where it appears.
- `world-orientation`: the compass's agreement with the map and what the clock shows.

### Modified Capabilities
- `hud-chrome`: menu buttons whose windows now exist are enabled; the rest say why they are greyed.

## Impact

- Possibly changed: `internal/server/protocol` and `internal/server/world` (stats for the windows; aggro rules), `web/chrome.js`, a new `web/sheet.js`, `docs/protocol.md`, tests.
- Depends on `stock-ui` and `stock-ui-rest` (both archived into the main specs).
- Art is EA-owned and stays in the gitignored `web/data/ui/`.
