// API client for horse. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { HorseResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the H.O.R.S.E. /horse/exec endpoint.
 *
 * `action` takes the betting verb of whichever discipline is running — they
 * share one vocabulary. **`bet` and `raise` require `amount`**; the server
 * rejects them without it rather than reading a missing amount as zero.
 */
export const horseApi = {
  exec: (
    command: 'reset' | 'action' | 'next' | 'hint' | 'log',
    params?: {
      action?: 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin';
      amount?: number;
      config?: { seats?: number; initialChips?: number; handsPerDiscipline?: number };
    },
  ) => gameExec<HorseResponse>('horse', { command, ...params }),
};
