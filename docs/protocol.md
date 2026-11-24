# Wire protocol

The browser and the game server talk JSON over one WebSocket (`/ws`), opened
with the session cookie from logging in. Every message is an object with a
`t` naming its type. Unknown types are ignored by both ends, so either side
can add messages without breaking an older peer; a change that would break
one raises `protocol.Version` (now 1), and a client that sends another is
refused with `{"t":"err","code":"version"}`.

The types live in `internal/server/protocol/protocol.go`, which is the
reference; this is the shape of the conversation.

## Joining

1. Client: `{"t":"hello","v":1,"char":<character id>}`, within ten seconds.
2. Server: `welcome` (your entity id, the zone, where you stand, which
   converted character to draw you with), then `inv` (pack and worn items)
   and `bar` (your styles), then a stream of `snap`s.
3. A character can be played from one place at a time. A second `hello` for
   it gets `{"t":"err","code":"join"}`, unless the first connection has gone and
   the character is lingering in a fight, in which case the newcomer takes it over.

## Movement

The client does not report where it is. It reports what it holds down, as
runs of fixed 1/60 s steps, and the server moves the character itself with the
same maths the browser predicts with (`web/move.js`, `move.go`):

```json
{"t":"input","seq":240,"steps":[[30,1,0,0,1],[30,1,0,1,1]],"px":31204.5,"py":29110.0}
```

Each run is `[count, fwd, strafe, turn, flags]`: that many steps with those
keys held (`fwd`/`strafe` -1..1, `turn` -1 left, 1 right, `flags` bit 0 = run, bit 1 = jump).
`seq` is the cumulative step count including these. `px`, `py` are the
client's predicted position, which the server only compares.

The server spends real time on steps: three per 50 ms tick, with a burst of at
most twenty banked, so a client that sends ten seconds of steps at once moves
one second's worth. Backing up and strafing are always at walking pace (85
units a second; running is 191). Collision (trees, walls, ledges) is part of
the step on both sides.

A slow scales every step's distance by the speed factor the server last sent
in `effects` (`speed`, 1 when free), and a held character (`speed` 0) neither
moves nor turns. The server applies its own factor when it steps the inputs,
so a client that ignores `effects` is corrected.

### Posture

The step also moves a body up and down: a jump starts with the flag bit 1 on the ground, a body in the air falls under gravity, ground that drops away by more than 12 units in a step makes it fall, and deep water holds it afloat at the surface. A client keeps stepping (with no keys) while it is in the air, and a body whose client stops sending is stepped down by the server after half a second. `you` carries the vertical state so the client's replay picks up where the server is: `air` (true in the air), `vz` (vertical speed), `pk` (the highest point of this stretch of air) and `jp` (the air began with a jump). A landing from a height costs hit points by the rule (`rules-falling`) and arrives as a `fall` event with `n` the damage and `x` the line.

Sitting, standing and sprinting are the chat commands `/sit` (a toggle), `/stand` and `/sprint` (a toggle). A player entity carries `ps`, its posture: `sit`, `air` or `swim`, absent when standing on the ground, and `vz` while in the air. `effects` has `spr` true while sprint is on: predict running steps at 1.25 times run speed.

The torch is the chat command `/torch` (a toggle; `/torch on` and `/torch off` set it). It is lit unless put out. A player entity carries `tc` true while its torch is lit, and `you` carries `tc` for your own; a dead player's torch is not lit. The client draws a lit torch as a light around its bearer after dusk, and always inside.

## Snapshots

Twenty a second:

```json
{"t":"snap","tick":812,"ack":240,
 "you":{"x":..,"y":..,"z":..,"h":yaw,"hp":175,"mhp":175,"end":88,"mend":100,"lv":1,
        "xp":8,"xpn":60,"mo":10,"tg":373,"at":true,"sp":191,"fg":true,"av":[0,2],"fix":false},
 "ents":[{"id":373,"k":"n","nm":"giant bull frog","m":100,"lv":1,"x":..,"y":..,"z":..,"h":..,
          "sp":0,"hp":0.82,"st":"fight"}],
 "ev":[{"k":"hit","from":1000001,"to":373,"n":9,"x":"Thor's Anvil"}]}
```

`ack` is how many input steps the server has applied; the client rewinds to
`you` at that step and replays the rest. `fix` means the prediction was wrong
enough to snap rather than ease. `ents` are the players (`k:"p"`, with `ch`,
the character to draw) and creatures (`k:"n"`, with `m`, the model number,
drawn as `data/char/m<N>`), plus `gd`, the guild or description shown under the name when it has one, and `eq`, the pieces it wears as `[slot, model, colour, extension]` (slots follow the database: 10 right hand, 11 left hand, 12 two-hand, 13 ranged, 21 helm, 22 gloves, 23 boots, 25 torso, 26 cloak, 27 legs, 28 arms). `eq` is sent once, on the snapshot where the creature enters range, and again if it leaves range and returns; keep it with the entity until the entity drops out of `ents`, within 14,000 units. `st` is `""`, `fight` or
`dead`. `ev` are the tick's events: `swing`, `hit`, `miss`, `block`, `parry`,
`evade`, `death`, `xp`, `level`, `loot`, `join`, and from the rule book:

- `xploss`: experience lost at death or on release, in `n`;
- `conloss`: constitution lost on release, in `n`;
- `pray`: a prayer at your gravestone, with `x` set to `started`, `cancelled` or `done` (`n` is the experience returned);
- `train`: a spec trained, with `x` the spec's name and `n` its new level.
- `spell`: a spell did damage or healed: `from` the caster, `to` the target, `x` the spell's name, `n` the damage (negative for a heal);
- `resist`: a hostile spell from `from` did not land on `to`; `x` is its name;
- `effect`: a lasting spell (damage over time, slow, stun, mesmerize) took hold on `to`; `x` is its name. A `block` event with `x` set is a bolt blocked.
- `cast`: a creature began casting: `from` the caster, `to` its target, `x` the spell, `dur` the cast time in seconds. As on a 1.65 server, a client shows the line and the caster's cast animation, not a cast bar;
- `fx` is added to the `cast`, `spell`, `effect` and `resist` events of a spell that has one: the number of the spell's row in the client's own spell table (`spells.csv` in `gamedata.mpk`), which says how the spell looks (its caster animations and effects, its projectile and its target effect). Absent when the spell has none or the world store was imported without it.
- `interrupt`: the caster `from` was interrupted (`x` the spell), and `castfail`: its spell failed because the target left its range or died.
- `draw`: the archer `from` began drawing a bow at `to`; `x` is `normal`, `critical` or `rapid`, `dur` the draw time in seconds. `drawstop` follows when the draw ends, by release or interruption. `arrow`: an arrow left `from` for `to` and takes `dur` seconds to arrive; the `hit`, `miss`, `block` or `evade` event for it follows at once. A shot is never parried.

A `level` event's `x` says what the level brought (stats, hit points, specialisation points), and it is also sent for a half-level from 40.

Gravestones are entities too, with `k:"g"`. `nm` is "grave of <name>" and `own` is true on your own. A gravestone can be targeted, and with your own targeted you can pray.

A creature entity also carries `cc`, its con colour to you: -3 grey, -2 green, -1 blue, 0 yellow, 1 orange, 2 red, 3 purple (rule book: `rules-experience`). It is absent on friendly NPCs. A client colours the creature's name by it.

Casters (Healer, Shaman, Runemaster) also have a spell power pool: `pw` and `mpw` on `you` in the snapshot, and `mpw` on `stats`; both are absent for a class with no power. Players carry their look too. Character creation (`POST /api/characters`) takes `female` (true for a female character; absent is male), and the character list marks female characters with `female:true`. `welcome` and each player entity have `rc` (the race number, 1 Briton, 2 Avalonian, 3 Highlander, 4 Saracen, 5 Norseman, 6 Troll, 7 Dwarf, 8 Kobold, 9 Celt, 10 Firbolg, 11 Elf, 12 Lurikeen, 13 Inconnu, 14 Valkyn, 15 Sylvan, 16 Half Ogre, 17 Frostalf, 18 Shar) and `fm` (true for a female character; absent means male). A client that has a converted base for that race and gender draws it and falls back to `ch` otherwise. A player entity's `eq` is in the same form as a creature's and is sent when the player comes into view and again whenever their gear changes; `gs:true` marks an `eq` that is present even when empty (the player took everything off), so keep the last one received until the next. Your own worn pieces arrive as `eq` on every `inv` message, which always has it, empty or not. Creation also takes `face` and `hair` (1 up; 0 or absent is the race's default look); the server refuses a face above 6 or a hair colour above 10 (4 for a Troll). `welcome` and a player entity carry them as `fc` and `hc` when chosen; `/api/me` archetypes carry their `race`. The client turns the numbers into skins with `data/appearance.json`. A client that does not know these fields ignores them.

## Requests

| type | body | does |
| --- | --- | --- |
| `target` | `id` | select a creature (0 clears); a corpse you may loot opens its window; another player can be targeted, to invite them, not attacked |
| `attack` | `on` | auto-attack on or off |
| `style` | `slot` | queue the style on that bar slot for the next swing |
| `release` | | a dead character stands up at their bind point now, without waiting for the respawn timer (5 s); ignored if alive |
| `loot` | `corpse`, `index` | take a corpse's item by index, or its money with -1 |
| `interact` | `id` | ask the creature with that id to do what it does: a merchant (`ik:"merchant"`) answers with `shop`, a trainer (`ik:"trainer"`) with `trainer` or a refusal if it teaches another class, a healer (`ik:"healer"`) with `healer` or a chat line if nothing is lost; a stable master (`ik:"stable"`) takes a horse ticket for one of its routes from the pack and starts the ride, otherwise answers with `shop` like a merchant (a ticket for another stable is refused in a chat line and kept); only within reach (128 units plus half the creature's reach); otherwise `err` |
| `train` | `spec`, `n` | raise a spec `n` levels (1 if absent) at a trainer of your class in reach; each level costs the level bought in specialisation points, and no spec goes above your level; answered with `trainer`, `stats` and `bar`, or `err` |
| `restore` | `npc` | pay the healer with that id, in reach, to restore as much lost constitution as your money covers; answered with `inv`, `stats` and a chat line, or `err` |
| `pray` | | pray at your targeted gravestone, within 2,000 units: after 5 seconds without moving or fighting you regain the experience it holds and it disappears. `/pray` in chat does the same |
| `buy` | `id` | buy one of the open shop's items by item id; the server checks stock, money, pack space and reach, then resends `inv` |
| `sell` | `slot` | sell the item in a pack slot (0-39) to the open shop for half its price, at least 1 copper; worn slots and worthless items are refused |
| `equip` | `from`, `to` | move between pack (0-39) and worn (40-47) slots; `to` -1 destroys, -2 wears it where it belongs. While a vault is open, either end may be a vault slot (1000-1039): items move between pack and vault, swap with what is there, and move within the vault; a worn item must be taken off first |
| `chat` | `ch`, `text` | `say` (3,000 units) or `all`; a leading `/` is a command (`/who`, `/pray`, `/bind`, `/gm ...`, the group commands, `/send` or `/tell <name> <text>` to one player of your realm in any zone, `/g <text>` to your group in any zone) |
| `groupop` | `op`, `id`, `name` | group requests: `invite` the player with `id` (in your zone) or `name` (anywhere, any case), `accept` or `decline` your invitation, `leave` the group, or, for the leader, `remove` the member `name`. Refusals come as `sys` chat lines. The same are typed as `/invite [name]` (your target without a name), `/accept`, `/decline`, `/disband` and `/remove <name>` |
| `gm` | `line` | an administrator's command: `goto`, `tp`, `level`, `heal`, `give`, `kill`, `where` |
| `gt` | `x`, `y`, `z`, `clear` | set the ground target for abilities aimed at the ground (Distraction): refused beyond 2,000 units or out of sight, with `err` `groundtarget`; `clear: true` drops it. The point comes back as `gt` in `you` (`[x, y, z]`); it is forgotten when you change zone |
| `tradeop` | `op`, `id`, `slot`, `cu` | a request in a trade: `open` with the player `id` (your target when 0; same zone and realm, within 512 units, neither in combat), `add` or `remove` a pack `slot`, set your offered money with `money` (`cu` copper), `accept` the offer as it stands, or `cancel`. `/trade` in chat opens one with your target. Refusals are `err` with `code` `trade` |
| `ping` | | answered with `pong` |

Everything a request asks is decided by the server: it can refuse (`err` with a
`code` and text for the player), and a client that claims a hit out of reach
deals no damage.

## Other server messages

`stats` is your character sheet, sent only to you: right after `welcome`, and again whenever any of it changes (a level, a trained spec, an item put on or taken off, constitution lost or restored). It carries:

- `cls`, `lv`, and `half` (past the midpoint of a level from 40);
- `base`, the eight stats chosen at creation, and `stats`, the same eight with level gains, items (capped at 1.5 × level) and lost constitution applied. Both are in the order strength, constitution, dexterity, quickness, intelligence, piety, empathy, charisma.
- `mhp` and `mend`;
- `specs`, rows of `key`, `nm`, `lv` and `bonus` (what items add, as combat counts it), and `pts`, the unspent specialisation points;
- `ab`, the abilities known, each `nm`, `lv` and `on`. `on` is false for an ability the server does not implement yet.
- `ws`, the held weapon's weapon skill, and `af`, the armour factor averaged over the body, both as shown;
- `conlost`, the constitution lost to deaths.

The numbers follow the rule book (`openspec/specs/rules-*`).

`effects` is your own lasting spells, sent only to you whenever they change: `list`, each with `nm` (the spell), `k` (what it does: `dot`, `slow`, `stun`, `mez`, `proc`, `buff`, `dmgadd`, `speed`, `regen`, `resist`, `sick`, `haste`, `shield`, `endregen`, `powregen`, `debuff`, `resdebuff`, `combatslow`, `disease`, `nearsight`, `fatigue`, `bladeturn`, `ablative`, `absorb`, `mezdur`, `stealth`, `camouflage`), `ic` (the imported icon number, which does not match this install's sheets: draw by `k`), `left` and `of` (seconds to run, and in all); `speed`, the share of normal speed you may move at (the strongest slow, 0 while held); and `held`, true while stunned or mesmerized, when you cannot move, swing or use styles. Predicted movement uses `speed` (see Movement). Effects end at death and are not kept over a logout.

The kinds also include `buff` (a stat or armour buff), `dmgadd` (a damage add), `speed` (a speed song), `regen` (a health regeneration song) and `resist` (a resist buff or chant). `song` is the id of the song or chant you have on (absent when none): casting a pulsing spell turns it on, it pulses on your group every few seconds, and casting it again turns it off. Only one song is on at a time, and a new one ends the first. `speed` can be above 1 while a speed song is on you; predicted movement applies it as it does a slow.

`book` is the spellbook, sent when it changes: `lines`, one per spec line, each with `spec`, `nm` and `sp`, its spells lowest level first. A spell has `id`, `nm`, `lv`, `pw` (power), `ct` (seconds to cast), `rg` (range), `tp` (the data's spell type), `tg` (its target), `ok` (false for a type the server cannot cast yet), `ic`, `rc` (recast seconds) and `pl` (true for a song or chant). A spell cast around the caster (range 0, an area) needs no target: send `cast` with `sp` (the spell `id`) and no `tg`.

`inv` (pack, worn items and money), `bar` (styles), `corpse` (a corpse's
loot), `shop`, `shopclose`, `chatmsg`, `err`, and `zone`: the character has moved to another zone,
so the client reloads at `?zone=N&play=<id>` and joins again. Crossing an
outdoor zone's edge, walking into an exit, and a GM `goto` all do this.

Each slot of `bar` carries `nm`, `cost`, `opens`, `lv`, the imported `icon`
number and `spec`, the style's weapon line ("Sword", "Hammer", "Axe"...). The
icon numbers come from a later client than the one a player installs, so the
web client draws a picture by `spec` and does not look `icon` up. Items in
`inv` and `corpse` carry `k` (weapon, shield, armour, misc), `w` (the worn slot
an armour piece belongs in) and `line` (a weapon's line) for the same reason.

`smith` is what a blacksmith (`ik:"smith"`) can repair for the player, sent when `interact` opens it: `npc`, `nm`, and `items`, each an `inv`-style row with `d` its condition and `pr` the repair price in copper. `repair` with a `slot` asks for a price and is answered with `repairquote` (`slot`, `nm`, `price`); `repair` with `accept` and the `price` they were quoted pays it, repairs the item to whole and sends a fresh `inv` and `smith` (or `err` if they cannot afford it; a changed price is quoted again). `smithclose` follows when they walk out of reach. Weapon, shield and armour rows in `inv` carry their condition in `d`.

`vault` is the character's own vault, sent when `interact` on a vault keeper (`ik:"vault"`) opens it, and again after each move: `npc`, `nm`, and `items` (slots 1000-1039). It goes only to its owner, with a fresh `inv` for the pack beside it. `vaultclose` follows when they walk out of reach, die or leave. The same vault opens at every keeper of the character's realm.

`shop` is a merchant's stock, sent when `interact` opens it: `npc` (its id), `nm`, and `items`, each an `inv`-style row (`id`, `nm`, `k`, `line`, `w`, `d`, `lv`) with `pr`, the price in copper, and `slot`, its place in the list in the order to show. Buy by `id`, not by `slot`. The shop stays open until `shopclose`, which the server sends when the player walks out of reach, dies or the merchant is gone; a character that changes zone simply starts with none open. A creature entity may carry `ik`, what interacting with it does (`"merchant"`, `"trainer"`, `"healer"`, `"stable"`); absent means nothing. A player entity carries `rd: true` while it rides a horse route: the server moves it along the route, its inputs do nothing, and creatures leave it alone. A ride across a zone border arrives as `zone`, as walking does, and the ride goes on in the next zone.

`trainer` is a trainer's offer: `npc`, `nm`, `pts` (unspent points) and `specs`, rows of `key`, `nm`, `lv`, `cost` (of the next level) and `can` (it can be trained now). `healer` is a healer's offer: `npc`, `nm`, `lost` (constitution lost), `price` (copper per point at your level) and `total`. Pictures for shop rows are chosen from `k` and `line`, like inventory rows; `m` and the imported icon numbers are not a reliable source.

`chatmsg` carries `ch`, the channel: `say`, `all`, `tell`, `group` or `sys` (the server speaking), with `from` and `text`. A tell you sent comes back to you with `to` (whom it went to) and no `from`. A tell to a name not in the game, or of another realm, is answered with a `sys` line "<name> is not in the game." The chat rate limit covers every channel.

`groupinv` is an invitation: `from` (the inviter's name) and `secs`, how long it stands (60). Answer with `groupop` `accept` or `decline`; unanswered it lapses, and the inviter is told it was declined.

`rez-offer` is a resurrection offered to you while you are dead: `from` (the caster's name), `health` and `power` (the percent of each you will have) and `secs`, how long it stands (15). Answer with `rez-answer` (client to server) and `accept` true or false; unanswered it lapses. `secs` 0 withdraws the offer (it lapsed, you released, or you answered). A raised character is stood up where the caster is, carries resurrection sickness (an effect of kind `sick`: 60 seconds, 30 percent less damage and spell strength) and skips the release stage of the death: no second experience loss, no constitution loss, no gravestone. A `rez` event (`from` the caster, `to` you, `x` the spell) is sent when it happens. A dead character may send `release` after 10 seconds, and is released after 15 minutes.

`conc` is the concentration you hold, sent only to casters whenever it changes: `max` (the pool), `used` (what your buffs hold) and `list`, each buff with `id`, `sp` (the spell id), `nm`, `tg` (the name of who it is on) and `cost`. A concentration buff has no timer: in `effects` it carries `cc` true and `left` and `of` 0. It ends when you cancel it (send `conc-cancel` with its `id`), when you die or leave, when the target dies or leaves, or when the target stays more than 5,000 units away for 20 seconds; the cost returns to the pool at once. The sheet's `stats` also carries `mcn` (the pool) and `cnu` (in use). A buff that does not fit is refused with `err` code `cast` and "This spell requires N concentration points to cast!"

`group` is your group as it stands, sent only to its members: when it forms or changes, and once a second while anything in it changes. `members` lists the leader first, then the rest in the order they joined, yourself included (`you`). Each has `name`, `level`, `class`, `hp` and `end` (percent), `zone` and `zoneName`, `inRange` (in your zone, within 6,000 units), `dead`, and `gone` while that member is out of the game. An empty `members` means you are in no group.

Limits: 16 KB a message, 120 messages a second per connection, 200
characters a chat line, one line a second with a burst of four.

`clock` is the game time of day: `gs`, seconds since midnight. It is sent after the welcome and every 30 seconds. A game day lasts 60 real minutes, so between messages the client runs it at 24 times real time. Night is 18:00 up to 06:00. Night-only and day-only creatures simply appear in and leave `ents` as their time begins and ends.

`barset` (`slot`, `id`) puts the style with that id, one on the `bar` the server sent, into a slot, swapping with the style there; the server answers with a fresh `bar`, or `err` if the character cannot use that style. `bar` may carry `also`: the styles the character has learned but cannot use with what they hold (`locked: true`, `slot: -1`), for the styles window. `stats` may carry `race`, the race's name. The bar order a player sets lasts until they reconnect.

### Pets

`pet` is your own pet, sent when it appears, changes or is lost: `id` (its entity id), `nm`, `lv`, `hp` (0 to 1), `st` (its stance: `passive`, `defensive` or `aggressive`), `md` (what it does: `follow`, `stay` or `attack`) and `mn`, the minions of a Bonedancer's commander (`id`, `nm`, `lv`, `hp`). `fx` is the pet's lasting spells, each as in `effects` (`nm`, `k`, `ic`, `left`, `of`); it is sent again when one starts or ends, not as the time runs down, so count `left` down from when the message arrived. The owner is also sent the `hit`, `miss`, `spell` and other events a pet or its minions give or take, at any distance. `cl` true means you no longer have one and the pet window closes. A pet is also a creature in `ents`, friendly, with `ow` set to its owner's name. Send `pet-cmd` with `cmd` set to `attack` (your current target), `follow`, `stay`, `release`, `passive`, `defensive` or `aggressive`; an order from someone who has no pet is answered with a system line. A pet is released when its owner dies, logs out, leaves the zone or stays more than 10,000 units away for 10 seconds, and a pet that dies leaves no corpse.

### Archery

Send `shoot` to use a bow (a bow in the ranged slot, arrows in a quiver slot). `mode` empty draws at the held target and keeps shooting until it dies, the arrows run out, the archer moves or you send `stop`; `critical` draws one Critical Shot; `rapid` turns Rapid Fire on or off (no shot); `stop` lowers the bow. A refused shot is answered with an `err` whose `code` is `shoot` ("You have no ammunition.", "That target is too far away!", and so on). The draw, release and hit are `draw`, `arrow` and the ordinary combat events (see Events). The ranged slot is inventory slot 48 and the four quiver slots are 49 to 52; arrows are sold in packs and stack. A draw stops when the archer moves or is hit.

### Stealth

A player with the Stealth spec hides and unhides with the chat command `/stealth` (or `/hide`). Hidden, the player has an effect of kind `stealth` (no end: `left` and `of` are 0) and `speed` in `effects` falls to the stealth speed, so predicted movement agrees. Other players are sent a hidden player in `ents` only while they are inside the detection range of the hider's Stealth spec against the viewer's level (a viewer with Detect Hidden sees further), and aggressive creatures take them as prey only inside theirs. Attacking, casting, shooting or taking damage ends stealth. Refusals are `err` with `code` `stealth`.

### Savage spells and timed abilities

A Savage's book (`book`) lists the Savagery line. Its self-buffs and endurance heal are paid in hit points: their entries carry `hp: true` and `pw` is the hit points the cast takes (a share of maximum health). A cast the character cannot pay is refused with `err` `cast` ("You don't have enough health."); the cost may not take them to 0. Their effects have kinds `physresist`, `dps`, `savevade` and `savparry` (the combat speed buff is `haste`). Berserk, Charge and QuickCast leave effects of kind `berserk`, `charge` and `quickcast`. The abilities are used with `/berserk`, `/charge`, `/quickcast` and `/distraction`; a refusal is a `sys` chat line.

### Trade

A trade is two players in one zone. Each change to either offer arrives to both as `trade` (`with`, and `mine` and `theirs`, each `{items, mo, ok}` with the items as in `inv` and `ok` set once that side has accepted) and clears both accepts; an offer is at most 10 items. When both have accepted, the server checks that each still holds what they offered and has pack room for what they get, then moves everything in one transaction: `tradeclose` with `done: true`, followed by `inv`. Otherwise nothing moves, and `tradeclose` carries `why` (not enough room, someone moved out of range, died, entered combat or cancelled).

### Duels, friends and following

`/duel` challenges your targeted player (same realm, within 1,500 units, neither in combat); they are sent `duelask` (`from`, `secs`) and answer `/accept` or `/decline` within 30 seconds. In a duel the two may attack each other (auto-attack, styles, hostile spells) and no one else may harm either; an entity in `ents` carries `foe: true` for the player you are dueling. A blow that would kill a duellist leaves them at 1 hit point and ends the duel with a zone-wide `sys` line; a duel gives no experience. It also ends with no winner when a creature strikes either, the two are over 2,500 units apart, or one leaves. Bows and pets do not take part in duels.

`/friend <name>` toggles a friend and `/ignore <name>` an ignore (both kept in the player store). `friends` (`list` of `{nm, on}`) is sent at login and whenever the list or a friend's presence changes; a friend coming or going is also a `sys` line. An ignored player's say, shout, emote and tells do not reach you. `/anon` hides your level and class from `/who`, `/lfg` marks you `[LFG]` there, `/afk [text]` answers tells with the text until you speak or move.

`/follow` walks you after your targeted player at run speed until you press a movement key or repeat the command; while it lasts `you.fl` is true, the server places you each tick and the client must not predict. `/assist` takes your target's target, `/face` turns you to your target, `/inspect` lists your targeted player's worn items as `sys` lines.

### Guilds

`/gc <form|invite|remove|leave|promote|demote|motd|omotd|note|emblem|ally|allyoff|info>` runs the guild commands; `/gu`, `/o` and `/as` talk on the guild, officer and alliance channels (chat messages with `ch` `guild`, `officer` or `alliance`). `form <name>` at a Guild Registrar, by the leader of a group of 8 of one realm, asks every member to `/accept`; the fee is taken from the leader when all agree. Invitations and alliance offers are answered with `/accept` and `/decline`. `guild` is sent at login and when anything about your guild changes: `nm` (empty when you have none), `rk` and `ti` (your rank and its title), `motd`, `omotd` (officers only), `em` (emblem number), `al` (the leading guild of your alliance), `pm` (what your rank may do, as bits: 1 guild speak, 2 guild hear, 4 officer speak, 8 officer hear, 16 invite, 32 promote, 64 remove, 128 emblem, 256 message) and `mem` (`{nm, rk, nt, on, lv, cl}`, level and class for members in the game). A player in a guild has `<Name>` in `gd` on their entity and in `/who`.

### Poisons

Send `use-item` with `slot` (a pack slot) to use an item; a poison item (`k` is `poison`) goes on the worn right-hand weapon when it has none, else the left, else replaces the right's. A poisoned weapon's `inv` item carries `ps`, the hits its poison has left, and says so in `d`. The poison fires on every hit and uses a charge; the last charge removes it ("Your weapon's poison is spent."). Refusals are `err` with `code` `use`.

### Tasks

`task` is your task, sent at login, when you take one, when it changes (a kill is marked, you finish it) and when it ends: `on` (false when you hold none), `k` (`kill` or `delivery`), `nm` ("Kill Task", "Delivery Task"), `tx` (what to do next), `gv` (who gave it), `dn` (a kill task whose creature is dead), `left` (seconds to the 2 hour limit), `n` (tasks finished at this level) and `max` (the most this level allows; tasks are for levels up to 20). Send `ask-task` with the entity `id` of an NPC to talk about tasks: a giver offers one (a named guard figure a kill task, a trainer, merchant or town crier a delivery task, half the time), the giver of a finished kill task pays, and the receiver of a delivery takes the item. A delivery task's item is in your pack (named for its receiver) and cannot be destroyed or stored. Send `abandon-task` to give it up; its item is removed and nothing is paid. The chat command `/task` says where you stand, `/task ask` talks to your target and `/task abandon` gives up. Talking to an NPC with `interact` also ends a task that ends at that NPC. Tasks and the count of tasks done are saved with the character, and the limit runs in real time.
