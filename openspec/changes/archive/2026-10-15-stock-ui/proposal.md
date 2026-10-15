## Why

The multiplayer client has a small custom UI (`web/ui.js`, `web/game.css`, about 260 lines) that looks nothing like the game and lacks a proper chat window, style bar, inventory or gear window. The install already holds the game's own UI as data: about 75 XML window layouts per skin, DDS and TGA atlases, icon sheets and bitmap fonts. Rendering from that data gives the real look and removes hand-drawn layout.

## What Changes

- Add `cmd/uiconv`, which reads the install's `ui/` XML, art and fonts and writes JSON layouts, PNG atlases and font sheets into the gitignored `web/data/ui/`.
- Add TGA and bitmap-font readers, since only DDS is supported today.
- Add a small JS renderer that builds windows from the converted layouts (window frames, buttons, image areas, text, icon sheets).
- Rebuild the HUD on it in this order: chat, target and health bars, style and quick bar, inventory and gear. Each stage is wired to the existing multiplayer protocol and replaces the matching part of `web/ui.js`.
- Spike first: find where the inventory/paperdoll and style-bar layouts live (not seen in `ui/*.xml`; candidates are `data/skin1/` and `icons/`, or hard-coded in the client). Where no layout exists, build the window by hand from the stock art and state that in the design.
- Converted art stays out of the repo, like every other client asset. **No protocol change.**

## Capabilities

### New Capabilities
- `ui-data-conversion`: turning the install's UI layouts, textures, icon sheets and fonts into web-loadable data.
- `stock-ui-windows`: building and drawing windows, buttons, text and icons from that data.
- `game-hud`: the in-game windows (chat, target and health, style and quick bar, inventory and gear) and how they act on the multiplayer protocol.

### Modified Capabilities
<!-- none: the multiplayer specs describe the server and protocol, which do not change -->

## Impact

- New: `cmd/uiconv`, `internal/ui` (XML, TGA, bitmap fonts), `web/uiwin.js` (renderer), HUD modules and tests.
- Replaced over time: `web/ui.js`, `web/game.css`; kept as the fallback when `web/data/ui/` is absent.
- `.gitignore` already covers `web/data/`; the README gains a UI conversion section.
- Art is EA-owned: converted locally, never committed, same as the rest of `web/data/`.
