// API client for baloot. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BalootConfig, BalootResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Baloot /baloot/exec endpoint. */
export const balootApi = {
  exec: (
    command: 'reset' | 'sun' | 'hokom' | 'pass' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<BalootConfig>,
    suit?: number,
  ) => gameExec<BalootResponse>('baloot', { command, cardIndex, config, suit }),
};
