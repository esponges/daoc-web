# Proposal: player-trade-and-social

## Why

Friends can group, chat and send tells, but they can't do the everyday things players did with each other in 1.65:

- **Hand over items or money.** There is no trade window. A drop for a groupmate is stuck in the looter's bag. The stock layout has a `Trade` window.
- **Duel.** `/duel` doesn't exist.
- **Keep a friends list.** The stock layout has `Friends` and `MiniFriends` windows, and `/friend` doesn't exist.
- **Use the rest of the social commands.** `/who` exists, but `/anon`, `/afk`, `/lfg`, `/ignore`, `/rp`, `/assist` and `/follow` don't.
- **Consider other players** or **inspect** them.
- **Use the stock Command window's FOLLOW and FACE buttons.** They are greyed.

These are the features that matter most when a handful of friends play together.

## What Changes

- **Trade:**
  - `/trade` with a targeted player in range opens the stock Trade window on both sides;
  - each side offers up to 10 items and money;
  - both must accept, and any change to the offer clears both accepts;
  - the server moves everything in one step, or nothing if either bag can't hold it.
- **Duel:**
  - `/duel` challenges, and the target accepts or declines;
  - both duellists fight with player-versus-player combat rules (introduced here and shared with `realm-versus-realm-combat`);
  - the duel ends at the first defeat, which leaves the loser at 1 hit point instead of dying;
  - a duel gives no experience and costs nothing, and ends if a duellist moves too far away or anyone else attacks.
- **Friends:** a per-character friends list kept on the server, with online status, in the stock Friends and MiniFriends windows. `/friend <name>` adds or removes.
- **Social commands:**
  - `/anon` hides the player's zone and level from `/who`;
  - `/afk [message]` replies to tells automatically;
  - `/lfg` flags the player in `/who`;
  - `/ignore` hides another player's chat and tells;
  - `/assist` takes the target's target;
  - `/follow` walks behind the target until it is cancelled;
  - `/face` turns to the target.
- **Inspect:** `/inspect` or the context menu shows another player's worn gear.
- **Command window:** FOLLOW and FACE are enabled.

## Capabilities

### New Capabilities
- `player-trade`: the trade window, offers, accepts and the transfer done in one step.
- `duels`: challenge, the duel's rules and its end.
- `friends-and-social`: the friends list, anon, afk, lfg, ignore, assist, follow, face and inspect.

### Modified Capabilities
- None as separate deltas:
  - player-versus-player combat between duellists is specified in `duels`, and `realm-versus-realm-combat` generalises it;
  - ignore and afk are specified in `friends-and-social`.

## Impact

- **Server:**
  - new `world/trade.go`, `world/duel.go` and `world/social.go`;
  - `players.db` gains a friends and ignore table (a migration);
  - combat resolves against players.
- **Protocol (additive):** trade messages, duel prompts, friends list, and flags in `/who`.
- **Client:** the stock Trade, Friends and MiniFriends windows from their templates, with plain fallbacks. Follow and face in the movement input.
