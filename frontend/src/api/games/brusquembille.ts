// API client for brusquembille. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BrusquembilleConfig, BrusquembilleResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Brusquembille /brusquembille/exec endpoint. */
export const brusquembilleApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<BrusquembilleConfig>,
  ) => gameExec<BrusquembilleResponse>('brusquembille', { command, cardIndex, config }),
};
