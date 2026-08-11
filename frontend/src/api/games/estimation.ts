// API client for estimation. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { EstimationConfig, EstimationResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Estimation /estimation/exec endpoint. */
export const estimationApi = {
  exec: (
    command: 'reset' | 'trump' | 'bid' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<EstimationConfig>,
    suit?: number,
    bid?: number,
  ) => gameExec<EstimationResponse>('estimation', { command, cardIndex, config, suit, bid }),
};
