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

## What was seen that led here

Recorded so the work can start from evidence, not memory.

- **Danger at the spawn.** At the Myrkwood Forest spawn (cell about 177, 99) a level 1 Troll Warrior with 175 hit points was killed by creatures labelled `giant bull frog 26` and `hog-nose slither 26`, standing a few steps away beside the lake, with hits of 60 to 98. Level 1 frogs stand among them, so a player cannot tell which a click will pick. The server aggroes any spawn with both `aggro_level` and `aggro_range` above zero; whether those values or the levels are the cause is the spike's question.
- **Loot.** Seen only with a sample corpse in the live page: the stock frame, icons by kind, the row click and TAKE ALL sending slots `[1]` then `[2, 1, 0, -1]`. A real kill and corpse still has to be watched end to end.
- **Release.** Seen only in the live page with a stand-in window: the timer bar and text, RELEASE and RELEASE BIND sending the release, CITY and HOUSE greyed, the window hiding on revival. The server side is covered by `TestPlayerCanReleaseEarly`. A real death still has to be watched.
- **Compass.** It turns clockwise with the camera (right gives E, SE, S, SW, W; left gives NW, N). Its north is the viewer's -y direction by assumption.
- **Clock.** Local time; the viewer has no day and night.
- **Menu bar.** Only INV. does anything. JOURNAL, MAP, WARMAP, ATTRIB., SKILLS, SPELLS, STYLES, GROUP and MENU are greyed.
- **Icons.** `Two Fists` has icon 105 and the first fist in the combat sheets is at 636; `Bloodletter` has 60 and the blood drop is at 505. Of 17,616 weapon and armour rows, 1,522 have a `model` in a range the install has a sheet for.
- **Driving the browser.** A held movement key cannot be sent through the automation's key action; the page's own key events, dispatched from script with a delay, move the character. The proofs in section 4 need that, and a safe creature.
