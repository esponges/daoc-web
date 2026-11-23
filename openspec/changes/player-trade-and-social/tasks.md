# Tasks

## 1. Trade

- [x] 1.1 `world/trade.go`: open, offer item or money, accept, cancel, and the single-step commit (D2). Tests per `player-trade` scenario.
- [x] 1.2 Protocol: the trade messages and the offer state sent to both sides. Update `docs/protocol.md` and add a protocol test.
- [x] 1.3 Client: the trade window (the stock layout has no Trade window to build from, so it is a plain panel), a click on a pack item adds it, the money field, accept and cancel.

## 2. Player-versus-player and duels

- [x] 2.1 The `mayHarm` predicate on the attack paths a duel uses (D3): melee, styles and hostile spells. Bows and pets do not fight players and take no part in duels.
- [x] 2.2 `world/duel.go`: a challenge answered with /accept and /decline (ahead of group invites), the end conditions, and the 1-health defeat. Tests per `duels` scenario.
- [x] 2.3 Client: the duel prompt (Accept and Decline), and the opponent's name in the enemy colour while the duel lasts.

## 3. Social

- [x] 3.1 `players.db` migration for friends and ignore. `/friend` and `/ignore`, login and logout notices.
- [x] 3.2 `/anon`, `/afk`, `/lfg` and their effect on `/who`; the afk reply.
- [x] 3.3 `/assist`, `/face`, `/follow` (D5) and `/inspect`. Tests.
- [x] 3.4 Client: a friends window (O; the stock Friends window is not built, so it is a plain panel); Command window FOLLOW and FACE enabled; /inspect answers in the chat log.

## 4. Check

- [x] 4.1 In two browser tabs on a scratch server: traded an item, started a duel and fought it (hits, blocks, evades), followed, and listed a friend. The defeat, assist and money are covered by server tests, not played. The new windows are plain panels, so they have no stock frame to fall back from; the Command window's plain buttons are tested in node.
- [x] 4.2 README and CLAUDE.md: what you can do today.
