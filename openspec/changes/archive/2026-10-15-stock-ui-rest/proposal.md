## Why

The `stock-ui` change moved chat, the summary and experience bars, the style bar and the inventory to the game's own windows, but the player still meets plain custom pieces on screen: a developer info panel, a text combat log, a loot box, a death banner, name labels in browser fonts, and a zone picker. Together they make the game read as a prototype wearing a skin.

## What Changes

- **Spike first.** For each remaining piece, record whether the install has a layout (`Release`, `menu_bar`, `Compass`, `Clock`, `Merchant`, `Vault` and `Interact` windows exist; a loot window and a combat log do not) or only art.
- Combat and system lines move into the stock chat window, in channel colours and on its Main tab, and the separate log goes.
- The loot window is rebuilt on a stock window's art with item pictures by kind and a Take all button.
- Death shows the stock Release window with its release button, replacing the banner.
- The developer info panel is hidden behind a key and the always-on key help goes; the stock menu bar, compass and clock take their place.
- Name labels, the zone picker and the remaining page text use the stock fonts and colours.
- No protocol change unless the spike finds one is needed. The built-in UI stays as the fallback when `web/data/ui/` is absent; converted art stays gitignored.

## Capabilities

### New Capabilities
- `game-messages`: where combat, loot, experience and system lines appear, and how they are coloured.
- `loot-and-release`: taking loot from a corpse and releasing after death, in stock windows.
- `hud-chrome`: the menu bar, compass, clock, hidden developer panel, zone picker, name labels and fonts around the game view.

### Modified Capabilities
<!-- none: game-hud lives in the stock-ui change and is not in the main specs yet -->

## Impact

- Changed: `web/ui.js`, `web/viewer.js`, `web/index.html`, `web/game.css`, `web/hud.js`, `web/uiwin.js` (compass and menu-bar controls).
- New: `web/loot.js`, `web/chrome.js`, tests for the pieces that can run headless.
- Depends on `stock-ui` (archive it first or build on top of it).
- Art is EA-owned and stays in the gitignored `web/data/ui/`.
