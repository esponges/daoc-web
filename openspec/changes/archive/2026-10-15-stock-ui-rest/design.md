## Context

`stock-ui` built the renderer (`web/uiwin.js`), the stock chat, summary, experience, style bar and inventory (`web/hud.js`, `web/pack.js`) and the converter (`cmd/uiconv`). What is still custom, by reading the viewer: the `#hud` info panel and key help in `index.html`, the `#log` combat log and `#err` line, the loot window and death banner in `ui.js`, name labels drawn as DOM elements in `viewer.js`, and the `#zones` picker. The install's layouts include `Release` (285x96: background, a timer bar, a RELEASE button and variants), `menu_bar` (800x27: MENU, JOURNAL, GROUP, STYLES, SPELLS, MAP, ATTRIB., INV. and others), `Compass` (100x100, a compass control), `Clock` (120x40, a time-of-day label), `Merchant` and `Vault` (large list windows over `dialog_*` frames) and `Interact`. There is no loot window and no combat log in the data; the real client used the chat window for combat lines.

## Goals / Non-Goals

**Goals:** nothing on screen in the default session looks like a prototype; every piece falls back to the built-in one when the converted data is absent.
**Non-Goals:** the journal, warmap, spells, skills and attribute windows the menu bar names; group windows; merchants; changing the protocol.

## Decisions

- **Combat lines go to chat, on the Main tab.** The stock chat window defines tabs (Main, Broad, Guild, Group, Chat). Game messages and speech are filed by kind; the viewer's `log(text, kind)` calls are routed to the chat window instead of `#log`. One call site, no new messages.
- **Loot is a small window on stock frames.** There is no layout for it, so it is built like the inventory: stock frame templates (`dlg_background_noresize`, `dialog_inset`) for the shell, rows of icon, name and detail drawn with stock fonts, and a Take all button from the button templates. Item pictures come from the kind-based icons of `stock-ui`.
- **Release uses the real window.** Build `Release` from its layout; the RELEASE button sends what the built-in banner sends today. The server's respawn is unchanged.
- **Chrome from layout.** The menu bar, compass and clock are built from their layouts. The compass needs a control the renderer lacks (`CompassControlTemplate` with a texture strip and the heading); the clock's time of day comes from the zone's sky time already in the viewer. Menu buttons that have no window yet are drawn disabled.
- **Developer panel behind a key.** One key toggles it; it starts hidden. This is a viewer change, not a UI-data one, so it applies with and without converted data.
- **Fonts.** Labels over heads switch from a CSS font to the bitmap font text builder where the cost is acceptable (up to 24 labels), else to the TrueType stock font. The login page loads the stock TrueType font from the same data when present.

## Risks / Trade-offs

- A per-frame bitmap-glyph label for 24 creatures may be slow in the DOM -> measure in the spike and fall back to the TrueType font.
- The menu bar names windows that do not exist yet -> draw them disabled rather than hide the bar's look.
- Chat tabs need per-tab filtering the current window lacks -> a simple kind filter, one buffer.

## Open Questions

- Does the compass control need a texture strip the converter does not write yet? Settled by task 1.1.

## Spike findings (task 1.1)

| Piece | Layout in the install | Art |
| --- | --- | --- |
| Death banner | `Release` window, 285x96: `normal_bg` frame, a timer bar (`release_timer_time`), a text line (`release_timer_text`, arial11) and four `generic_hr_button`s: /RELEASE, /RELEASE BIND, /RELEASE CITY, /RELEASE HOUSE | atlantis atlases |
| Menu bar | `menu_bar`, 800x27: one strip holding the experience bubbles and bar (so the separate experience window is redundant when it is up), the money (four labels with coin decals), a MENU button and buttons JOURNAL, WARMAP, MAP, ATTRIB., INV., SKILLS, SPELLS, STYLES, GROUP | atlantis atlases |
| Compass | `Compass`, 100x100: a `CompassControlDef` bound to `compass_heading`, 128x32 or 64x16, over `compass` = `isles/compass.bmp`, a 512x32 strip | already converted |
| Clock | `Clock`, 120x40: frame and a `time_of_day` label (arial11) | atlantis atlases |
| Loot | none (the closest are `Merchant`, 350x458, and `Vault`, 300x430, list windows over `dialog_background_*` and `dialog_inset` frames) | `skin1/merchant.bmp`, `vault.bmp` |
| Combat log | none: the real client printed combat in the chat window | n/a |
| Info panel, zone picker, name labels | none: a developer viewer's own | n/a |

So the Release, Clock and menu bar windows are layout-driven and need only the existing renderer plus the bits it lacks (a compass control, a button label from the layout). The compass is a strip scrolled by heading, wrapping every 512 pixels. Loot has no layout and is built on frames. Nothing here needs a converter change. The menu bar already draws the experience and money, so the stock-ui experience window can go once the bar is in.

## What was decided while building

- **The clock shows the player's own time.** The viewer has no day and night, so there is no game time to show; inventing a day length would be fiction. The stock clock window shows local time.
- **The compass assumes north is the world's -y direction**, the viewer's own heading convention (`angleTo`). It turns the right way with the camera; whether the labels match the real map is not checked against the install's data.
- **A `release` message was added** (additive). The server stood a dead character up after five seconds by itself, so a RELEASE button needed something to send. RELEASE and RELEASE BIND both send it, since the bind point is the only place the server returns a character to; RELEASE CITY and RELEASE HOUSE are greyed.
- **The menu bar replaces the separate experience window.** It already carries the experience bubbles, the money and the window buttons; the INV. button opens the inventory and the other buttons, for windows that do not exist, are greyed.
- **The developer panel is a viewer concern**: hidden when playing, toggled with the backquote key, up by default in the offline viewer.
