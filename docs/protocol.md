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
keys held (`fwd`/`strafe` -1..1, `turn` -1 left, 1 right, `flags` bit 0 = run).
`seq` is the cumulative step count including these. `px`, `py` are the
client's predicted position, which the server only compares.

The server spends real time on steps: three per 50 ms tick, with a burst of at
most twenty banked, so a client that sends ten seconds of steps at once moves
one second's worth. Backing up and strafing are always at walking pace (85
units a second; running is 191). Collision (trees, walls, ledges) is part of
the step on both sides.

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

A `level` event's `x` says what the level brought (stats, hit points, specialisation points), and it is also sent for a half-level from 40.

Gravestones are entities too, with `k:"g"`. `nm` is "grave of <name>" and `own` is true on your own. A gravestone can be targeted, and with your own targeted you can pray.

A creature entity also carries `cc`, its con colour to you: -3 grey, -2 green, -1 blue, 0 yellow, 1 orange, 2 red, 3 purple (rule book: `rules-experience`). It is absent on friendly NPCs. A client colours the creature's name by it.

Players carry their look too. `welcome` and each player entity have `rc` (the race number, 1 Briton, 2 Avalonian, 3 Highlander, 4 Saracen, 5 Norseman, 6 Troll, 7 Dwarf, 8 Kobold, 9 Celt, 10 Firbolg, 11 Elf, 12 Lurikeen, 13 Inconnu, 14 Valkyn, 15 Sylvan, 16 Half Ogre, 17 Frostalf, 18 Shar) and `fm` (true for a female character; absent means male). A client that has a converted base for that race and gender draws it and falls back to `ch` otherwise. A player entity's `eq` is in the same form as a creature's and is sent when the player comes into view and again whenever their gear changes; `gs:true` marks an `eq` that is present even when empty (the player took everything off), so keep the last one received until the next. Your own worn pieces arrive as `eq` on every `inv` message, which always has it, empty or not. A client that does not know these fields ignores them.

## Requests

| type | body | does |
| --- | --- | --- |
| `target` | `id` | select a creature (0 clears); a corpse you may loot opens its window |
| `attack` | `on` | auto-attack on or off |
| `style` | `slot` | queue the style on that bar slot for the next swing |
| `release` | | a dead character stands up at their bind point now, without waiting for the respawn timer (5 s); ignored if alive |
| `loot` | `corpse`, `index` | take a corpse's item by index, or its money with -1 |
| `interact` | `id` | ask the creature with that id to do what it does: a merchant (`ik:"merchant"`) answers with `shop`, a trainer (`ik:"trainer"`) with `trainer` or a refusal if it teaches another class, a healer (`ik:"healer"`) with `healer` or a chat line if nothing is lost; only within reach (128 units plus half the creature's reach); otherwise `err` |
| `train` | `spec`, `n` | raise a spec `n` levels (1 if absent) at a trainer of your class in reach; each level costs the level bought in specialisation points, and no spec goes above your level; answered with `trainer`, `stats` and `bar`, or `err` |
| `restore` | `npc` | pay the healer with that id, in reach, to restore as much lost constitution as your money covers; answered with `inv`, `stats` and a chat line, or `err` |
| `pray` | | pray at your targeted gravestone, within 2,000 units: after 5 seconds without moving or fighting you regain the experience it holds and it disappears. `/pray` in chat does the same |
| `buy` | `id` | buy one of the open shop's items by item id; the server checks stock, money, pack space and reach, then resends `inv` |
| `sell` | `slot` | sell the item in a pack slot (0-39) to the open shop for half its price, at least 1 copper; worn slots and worthless items are refused |
| `equip` | `from`, `to` | move between pack (0-39) and worn (40-47) slots; `to` -1 destroys, -2 wears it where it belongs |
| `chat` | `ch`, `text` | `say` (3,000 units) or `all`; a leading `/` is a command (`/who`, `/gm ...`) |
| `gm` | `line` | an administrator's command: `goto`, `tp`, `level`, `heal`, `give`, `kill`, `where` |
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

`shop` is a merchant's stock, sent when `interact` opens it: `npc` (its id), `nm`, and `items`, each an `inv`-style row (`id`, `nm`, `k`, `line`, `w`, `d`, `lv`) with `pr`, the price in copper, and `slot`, its place in the list in the order to show. Buy by `id`, not by `slot`. The shop stays open until `shopclose`, which the server sends when the player walks out of reach, dies or the merchant is gone; a character that changes zone simply starts with none open. A creature entity may carry `ik`, what interacting with it does (`"merchant"`, `"trainer"`, `"healer"`); absent means nothing.

`trainer` is a trainer's offer: `npc`, `nm`, `pts` (unspent points) and `specs`, rows of `key`, `nm`, `lv`, `cost` (of the next level) and `can` (it can be trained now). `healer` is a healer's offer: `npc`, `nm`, `lost` (constitution lost), `price` (copper per point at your level) and `total`. Pictures for shop rows are chosen from `k` and `line`, like inventory rows; `m` and the imported icon numbers are not a reliable source.

Limits: 16 KB a message, 120 messages a second per connection, 200
characters a chat line, one line a second with a burst of four.
