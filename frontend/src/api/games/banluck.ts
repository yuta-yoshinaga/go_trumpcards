// API client for banluck. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BanLuckResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Ban Luck /banluck/exec endpoint.
 *
 * **`bet` must be sent even on your banker round**, where the correct value is
 * `0` — the banker does not stake. Omitting it is an error, not a way to say
 * "no bet", so the server rejects a `bet` command with no amount.
 */
export const banluckApi = {
  exec: (command: 'reset' | 'bet' | 'hit' | 'stand' | 'next' | 'hint' | 'log', params?: { bet?: number }) =>
    gameExec<BanLuckResponse>('banluck', { command, ...params }),
};
