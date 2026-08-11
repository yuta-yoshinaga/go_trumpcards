// API client for reversis. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ReversisConfig, ReversisResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Reversis /reversis/exec endpoint. */
export const reversisApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<ReversisConfig>,
  ) => gameExec<ReversisResponse>('reversis', { command, cardIndex, config }),
};
