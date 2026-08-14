// API client for polignac. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PolignacConfig, PolignacResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Polignac /polignac/exec endpoint. */
export const polignacApi = {
  exec: (
    command: 'reset' | 'capot' | 'pass' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<PolignacConfig>,
  ) => gameExec<PolignacResponse>('polignac', { command, cardIndex, config }),
};
