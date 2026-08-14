// API client for rams. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RamsConfig, RamsResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Rams /rams/exec endpoint. */
export const ramsApi = {
  exec: (
    command: 'reset' | 'in' | 'out' | 'card' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<RamsConfig>,
  ) => gameExec<RamsResponse>('rams', { command, cardIndex, config }),
};
