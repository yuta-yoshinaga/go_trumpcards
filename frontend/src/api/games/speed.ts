// API client for speed. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SpeedConfig, SpeedResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Speed /speed/exec endpoint. */
export const speedApi = {
  exec: (
    command: 'reset' | 'play' | 'flip' | 'hint' | 'log',
    cardIndex?: number,
    pileIndex?: number,
    config?: SpeedConfig,
  ) => gameExec<SpeedResponse>('speed', { command, cardIndex, pileIndex, ...config }),
};
