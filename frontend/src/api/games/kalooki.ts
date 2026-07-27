// API client for kalooki. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KalookiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Kalooki /kalooki/exec endpoint. */
export const kalookiApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'layoff' | 'discard' | 'nextround' | 'log',
    params?: {
      cardIndex?: number;
      meldGroups?: number[][];
      targetPlayerIdx?: number;
      meldIdx?: number;
      config?: { cpuDifficulty?: number; playerCount?: number; openingThreshold?: number };
    },
  ) => gameExec<KalookiResponse>('kalooki', { command, ...(params ?? {}) }),
};
