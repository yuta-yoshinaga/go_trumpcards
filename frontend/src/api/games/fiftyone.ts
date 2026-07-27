// API client for fiftyone. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FiftyOneResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Fifty-one /fiftyone/exec endpoint. */
export const fiftyoneApi = {
  exec: (
    command: 'reset' | 'play' | 'exchangeall' | 'stop' | 'log',
    opts?: { handIdx?: number; tableIdx?: number; config?: { cpuDifficulty?: number } },
  ) => gameExec<FiftyOneResponse>('fiftyone', { command, ...opts }),
};
