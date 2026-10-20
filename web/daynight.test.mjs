// node web/daynight.test.mjs — the client's game clock and how dark the scene is.
import { gameSeconds, isNight, nightness, clockFace, nightLook, DAY } from './daynight.js';

let failed = 0;
const ok = (cond, msg) => { if (!cond) { failed++; console.error('FAIL: ' + msg); } else console.log('ok: ' + msg); };

const clock = { gs: 17 * 3600 + 59 * 60, at: 100 };
ok(gameSeconds(clock, 100) === 17 * 3600 + 59 * 60, 'the clock reads what the server sent');
ok(gameSeconds(clock, 102.5) === 17 * 3600 + 59 * 60 + 60, '2.5 real seconds are one game minute');
ok(gameSeconds({ gs: DAY - 10, at: 0 }, 1) === 14, 'the clock wraps past midnight');
ok(gameSeconds(null, 5) === null, 'without a clock there is no game time');
ok(!isNight(17 * 3600 + 59 * 60 + 59) && isNight(18 * 3600) && isNight(0) && isNight(6 * 3600 - 1) && !isNight(6 * 3600), 'night is 18:00 up to 06:00');
ok(nightness(18 * 3600) === 0 && Math.abs(nightness(18.5 * 3600) - 0.5) < 1e-9 && nightness(19 * 3600) === 1, 'it darkens over the game hour after dusk');
ok(nightness(0) === 1 && nightness(5 * 3600) === 1 && Math.abs(nightness(5.5 * 3600) - 0.5) < 1e-9 && nightness(6 * 3600) === 0 && nightness(12 * 3600) === 0, 'it lightens over the hour before dawn');
const face = clockFace(18 * 3600 + 5 * 60 + 30);
ok(face.getHours() === 18 && face.getMinutes() === 5, 'the clock face shows hours and minutes');
const day = { fog: [0.5, 0.6, 0.7], ambient: [0.4, 0.4, 0.4], sun: [0.9, 0.9, 0.8] };
ok(nightLook(day, 0).fog.every((v, i) => v === day.fog[i]), 'by day the look is the day look');
const dark = nightLook(day, 1);
ok(dark.fog[2] < day.fog[2] && dark.ambient[0] < day.ambient[0] && dark.sun[0] < 0.1, 'at night fog, ambient and sun are darker');
process.exit(failed ? 1 : 0);
