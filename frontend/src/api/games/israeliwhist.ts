// API client for israeliwhist. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { IsraeliWhistConfig, IsraeliWhistResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Israeli Whist /israeliwhist/exec endpoint. */
export const israeliwhistApi = {
  exec: (
    command: 'reset' | 'auction' | 'pass' | 'bid' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<IsraeliWhistConfig>,
    suit?: number,
    bid?: number,
  ) => gameExec<IsraeliWhistResponse>('israeliwhist', { command, cardIndex, config, suit, bid }),
};
