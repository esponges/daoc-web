// node web/vault.test.mjs — the vault window's cells and where an item goes.
import { vaultCells, firstFreeVault, firstFreePack, VAULT_BASE, VAULT_SLOTS } from './vault.js';

let failed = 0;
const ok = (cond, msg) => { if (!cond) { failed++; console.error('FAIL: ' + msg); } else console.log('ok: ' + msg); };

const vault = { t: 'vault', npc: 9, nm: 'Keeper', items: [{ slot: VAULT_BASE, nm: 'hammer' }, { slot: VAULT_BASE + 2, nm: 'shield', n: 1 }] };
const cells = vaultCells(vault);
ok(cells.length === VAULT_SLOTS && cells[0].item.nm === 'hammer' && !cells[1].item && cells[2].item.nm === 'shield', 'forty cells with the items in their slots');
ok(firstFreeVault(vault) === VAULT_BASE + 1, 'the first free vault slot is the first gap');
ok(firstFreeVault({ items: cells.map((c) => ({ slot: c.slot })) }) === -1, 'a full vault has no free slot');
ok(firstFreeVault(null) === VAULT_BASE && vaultCells(null).every((c) => !c.item), 'no vault message is an empty vault');
ok(firstFreePack({ items: [{ slot: 0 }, { slot: 1 }, { slot: 3 }] }) === 2, 'the first free pack slot is the first gap');
ok(firstFreePack({ items: Array.from({ length: 40 }, (_, s) => ({ slot: s })) }) === -1, 'a full pack has no free slot');
ok(firstFreePack({ items: [{ slot: 40 }] }) === 0, 'worn slots do not count as pack slots');
process.exit(failed ? 1 : 0);
