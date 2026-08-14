// API client for pig. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PigConfig, PigResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Pig /pig/exec endpoint.
 *
 * **`signal` and `next` take no card.** Noticing a signal is a different action
 * from passing, not `pass` with the index left out.
 */
export const pigApi = {
  exec: (
    command: 'reset' | 'pass' | 'signal' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<PigConfig>,
  ) => gameExec<PigResponse>('pig', { command, cardIndex, config }),
};
