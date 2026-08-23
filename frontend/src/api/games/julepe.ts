// API client for julepe. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { JulepeConfig, JulepeResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Julepe /julepe/exec endpoint. */
export const julepeApi = {
  exec: (
    command: 'reset' | 'in' | 'out' | 'card' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<JulepeConfig>,
  ) => gameExec<JulepeResponse>('julepe', { command, cardIndex, config }),
};
