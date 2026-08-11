// API client for hokm. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { HokmConfig, HokmResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Hokm /hokm/exec endpoint. */
export const hokmApi = {
  exec: (
    command: 'reset' | 'trump' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<HokmConfig>,
    suit?: number,
  ) => gameExec<HokmResponse>('hokm', { command, cardIndex, config, suit }),
};
