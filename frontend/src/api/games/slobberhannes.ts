// API client for slobberhannes. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SlobberhannesConfig, SlobberhannesResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Slobberhannes /slobberhannes/exec endpoint. */
export const slobberhannesApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<SlobberhannesConfig>,
  ) => gameExec<SlobberhannesResponse>('slobberhannes', { command, cardIndex, config }),
};
