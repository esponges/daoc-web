## Context

After `stock-ui` and `stock-ui-rest`, the stock windows cover chat, summary, style bar, inventory, loot, release, menu bar, compass and clock. What the server sends a client today is health, endurance, level, experience and money for the character, entities in snapshots, the style bar, the inventory and corpses; it sends no attributes, skills, spell lists or groups. The install has layouts for the windows the menu bar names: `StatsAttributes` (192x416), `Stats_Spec_Abil` (192x416), `StatsGroup` (192x416), `new_group` (300x150), `MiniGroup` (128x116), and art for the classic character sheet in `data/skin1/charsheet.bmp`. The imported creature data carries `aggro_level` and `aggro_range` per spawn; the server aggroes a creature that has both greater than zero (`npc.go`). Creatures of level 26 stand beside the Myrkwood spawn and attack a level 1 character in two blows.

## Goals / Non-Goals

**Goals:** the greyed menu windows exist where the server has the data; a new character is not killed at once by what stands next to it; the two flows still unseen are seen; the compass and clock stop being guesses.
**Non-Goals:** groups, quests and the journal as game systems; new chat channels; new character classes.

## Decisions

- **Windows follow data.** A window is built only if the server can fill it. The attributes window needs the six attributes, level, race and class; the server holds class and level and would need to send attributes (additive, owner only). Skills, spells and styles windows list what the imported tables give a Warrior (specialisations and styles already drive the style bar). Anything the server cannot fill stays blank, never invented.
- **Group stays greyed** unless the spike finds a group system to show; a window over nothing is the worst of both.
- **Safety is a rule, not an edit.** The imported data is never changed. The likely rule is that a spawn whose level is far above a starting zone's expected range is placed but made passive (no aggro) near the spawn point, or left out of a radius around it, reported at load. The spike decides whether the cause is data (an aggro level of zero meaning "always" in the source) or the server's reading of it.
- **Icons: one bounded look, then stop.** The earlier work showed the imported numbers belong to a later client. The follow-up checks whether any community table maps them to this client's sheets; if none, the kind-based pictures stay and the design says so.
- **Compass is checked, not assumed.** A zone exit's position from the zone data gives a known direction from a known point; the compass is compared against it.
- **Clock: game time only with evidence.** If the install says nothing about a day length, the clock keeps local time and its tooltip says so.

## Risks / Trade-offs

- Making creatures passive near spawn changes the game's feel in a starting zone -> the rule is limited to a radius around where characters appear and reported, and is easy to switch off.
- A new attributes message grows the protocol -> additive, owner-only, documented, and the window is built from what arrives.
- The real-kill and real-death proofs depend on reaching a safe creature -> done after the safety rule, which removes the danger.

## Open Questions

- Is the dangerous aggro in the data or in the server's reading of it? Settled by task 1.1.
- Does the server hold the attributes it would need to send? Settled by task 1.1.
