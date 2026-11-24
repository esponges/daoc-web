// Headless checks of the quest window's pure parts.
//   node web/quest.test.mjs
import { questView, offerView } from './quest.js';

let failed = 0;
const ok = (cond, name) => {
  if (!cond) { failed++; console.log(`FAIL ${name}`); } else console.log(`ok   ${name}`);
};

const v = questView({
  t: 'quests',
  active: [{ id: 'wolves', nm: 'Wolves at the Gate', text: 'Kill two wolves.', have: 1, need: 2 },
    { id: 'letter', nm: 'A Letter', text: 'Bring it to Svala.', have: 1, need: 1 }],
  done: ['Old Business'],
});
ok(v.rows.length === 2 && v.rows[0].title === 'Wolves at the Gate', 'the journal lists active quests');
ok(v.rows[0].text === 'Kill two wolves. (1 of 2)', 'a counted step shows its progress');
ok(v.rows[1].text === 'Bring it to Svala.', 'a single item step shows no count');
ok(v.done[0] === 'Old Business', 'done quests are listed by title');
ok(questView(null).rows.length === 0, 'no journal yet is empty');
const o = offerView({ id: 'wolves', nm: 'Wolves at the Gate', npc: 'Ragna', text: 'Thin them.', reward: '10% of a level' });
ok(o.head === 'Wolves at the Gate - Ragna' && o.reward === 'Reward: 10% of a level', 'an offer names the quest, the giver and the reward');
ok(offerView({ id: 'x', nm: 'X', npc: 'N', text: 't' }).reward === '', 'no reward text when there is none');
if (failed) process.exit(1);
