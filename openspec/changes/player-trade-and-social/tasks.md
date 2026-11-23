# Tasks

## 1. Trade

- [x] 1.1 `world/trade.go`: open, offer item or money, accept, cancel, and the single-step commit (D2). Tests per `player-trade` scenario.
- [x] 1.2 Protocol: the trade messages and the offer state sent to both sides. Update `docs/protocol.md` and add a protocol test.
- [ ] 1.3 Client: the stock Trade window from its template, drag from the bag, the money field, accept; a plain fallback.

## 2. Player-versus-player and duels

- [x] 2.1 The `mayHarm` predicate on the attack paths a duel uses (D3): melee, styles and hostile spells. Bows and pets do not fight players and take no part in duels.
- [x] 2.2 `world/duel.go`: a challenge answered with /accept and /decline (ahead of group invites), the end conditions, and the 1-health defeat. Tests per `duels` scenario.
- [ ] 2.3 Client: the duel prompt, and enemy-colour names for the opponent while the duel lasts.

## 3. Social

- [x] 3.1 `players.db` migration for friends and ignore. `/friend` and `/ignore`, login and logout notices.
- [x] 3.2 `/anon`, `/afk`, `/lfg` and their effect on `/who`; the afk reply.
- [x] 3.3 `/assist`, `/face`, `/follow` (D5) and `/inspect`. Tests.
- [ ] 3.4 Client: the stock Friends and MiniFriends windows; Command window FOLLOW and FACE enabled; the inspect view.

## 4. Check

- [ ] 4.1 In two browser tabs: trade an item and money, duel to a defeat, follow and assist. Check the fallback UI.
- [ ] 4.2 README and CLAUDE.md: what you can do today.
