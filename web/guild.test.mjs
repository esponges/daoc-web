// Headless checks of the guild window's pure parts.
//   node web/guild.test.mjs
import { guildView, can, PERM } from './guild.js';

let failed = 0;
const ok = (cond, name, extra = '') => {
  if (!cond) { failed++; console.log(`FAIL ${name} ${extra}`); } else console.log(`ok   ${name}`);
};

const m = {
  nm: 'Hammers of Odin', rk: 1, ti: 'Officer', motd: 'Raid on Friday', em: 12, al: 'Sons of Tyr', pm: 16 | 32 | 64 | 3,
  mem: [
    { nm: 'Sigrid', rk: 9, on: false },
    { nm: 'Thora', rk: 0, on: true, lv: 40, cl: 'Warrior' },
    { nm: 'Bjorn', rk: 1, on: true, lv: 25, cl: 'Healer', nt: 'healer' },
  ],
};
const v = guildView(m);
ok(v.title === '<Hammers of Odin>' && v.sub === 'Officer, rank 1', 'the title and your rank');
ok(v.rows.map((r) => r.name).join() === 'Thora,Bjorn,Sigrid', 'members sort by rank, then name');
ok(guildView(m, { sort: 'level' }).rows[0].name === 'Thora', 'they sort by level, highest first');
ok(guildView(m, { sort: 'name' }).rows.map((r) => r.name).join() === 'Bjorn,Sigrid,Thora', 'or by name');
ok(guildView(m, { showOffline: false }).rows.length === 2, 'offline members can be hidden');
ok(v.online === 2 && v.total === 3, 'online and total counts');
ok(v.rows[1].text.includes('L25 Healer') && v.rows[1].text.includes('- healer'), 'a row shows level, class and note');
ok(v.alliance === 'Alliance: Sons of Tyr' && v.motd === 'Raid on Friday' && v.emblem === 12, 'alliance, message and emblem');
ok(v.mayInvite && v.mayPromote && v.mayRemove && !v.mayMotd && !v.mayEmblem, 'the buttons follow the rank\'s permissions');
ok(can(3, PERM.guildSpeak) && !can(3, PERM.officerSpeak), 'permission bits');
ok(guildView({ t: 'guild', mem: [] }) === null && guildView(null) === null, 'no guild, no view');
if (failed) process.exit(1);
