// API client for shelem. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ShelemConfig, ShelemResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Shelem /shelem/exec endpoint. */
export const shelemApi = {
  exec: (
    command: 'reset' | 'bid' | 'shelem' | 'pass' | 'discard' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<ShelemConfig>,
    suit?: number,
    bid?: number,
    discards?: number[],
  ) => gameExec<ShelemResponse>('shelem', { command, cardIndex, config, suit, bid, discards }),
};
