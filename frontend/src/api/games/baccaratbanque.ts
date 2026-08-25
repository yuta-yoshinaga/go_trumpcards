// API client for baccaratbanque. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BaccaratBanqueResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Baccarat Banque /baccaratbanque/exec endpoint.
 *
 * **Drawing and standing are separate commands.** Folding them into one
 * command with a boolean lets a request that arrives without the flag fall
 * silently into whichever branch the default picks.
 */
export const baccaratbanqueApi = {
  exec: (
    command: 'reset' | 'draw' | 'stand' | 'nextcoup' | 'retire' | 'hint' | 'log',
    params?: { cpuDifficulty?: number; startChips?: number; betAmount?: number },
  ) =>
    gameExec<BaccaratBanqueResponse>('baccaratbanque', {
      command,
      ...(params ? { config: params } : {}),
    }),
};
