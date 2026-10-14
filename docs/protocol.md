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
drawn as `data/char/m<N>`) within 14,000 units. `st` is `""`, `fight` or
`dead`. `ev` are the tick's events: `swing`, `hit`, `miss`, `block`, `parry`,
`evade`, `death`, `xp`, `level`, `loot`, `join`.

## Requests

| type | body | does |
| --- | --- | --- |
| `target` | `id` | select a creature (0 clears); a corpse you may loot opens its window |
| `attack` | `on` | auto-attack on or off |
| `style` | `slot` | queue the style on that bar slot for the next swing |
| `loot` | `corpse`, `index` | take a corpse's item by index, or its money with -1 |
| `equip` | `from`, `to` | move between pack (0-39) and worn (40-47) slots; `to` -1 destroys, -2 wears it where it belongs |
| `chat` | `ch`, `text` | `say` (3,000 units) or `all`; a leading `/` is a command (`/who`, `/gm ...`) |
| `gm` | `line` | an administrator's command: `goto`, `tp`, `level`, `heal`, `give`, `kill`, `where` |
| `ping` | | answered with `pong` |

Everything a request asks is decided by the server: it can refuse (`err` with a
`code` and text for the player), and a client that claims a hit out of reach
deals no damage.

## Other server messages

`inv` (pack, worn items and money), `bar` (styles), `corpse` (a corpse's
loot), `chatmsg`, `err`, and `zone`: the character has moved to another zone,
so the client reloads at `?zone=N&play=<id>` and joins again. Crossing an
outdoor zone's edge, walking into an exit, and a GM `goto` all do this.

Limits: 16 KB a message, 120 messages a second per connection, 200
characters a chat line, one line a second with a burst of four.
