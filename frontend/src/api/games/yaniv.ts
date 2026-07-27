// API client for yaniv. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { YanivResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Yaniv game settings. */
export interface YanivConfigInput {
  cpuDifficulty?: number;
  scoreLimit?: number;
}

/** API client for the Yaniv /yaniv/exec endpoint. */
export const yanivApi = {
  exec: (
    command: 'reset' | 'discard' | 'yaniv' | 'drawstock' | 'drawpickup' | 'nextround' | 'log',
    opts?: { cardIndices?: number[]; end?: number; config?: YanivConfigInput },
  ) =>
    gameExec<YanivResponse>('yaniv', {
      command,
      cardIndices: opts?.cardIndices,
      end: opts?.end,
      config: opts?.config,
    }),
};
