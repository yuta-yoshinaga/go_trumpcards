// API client for carioca. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CariocaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Carioca /carioca/exec endpoint. */
export const cariocaApi = {
  exec: (
    command:
      | 'reset'
      | 'drawstock'
      | 'drawdiscard'
      | 'meldcontract'
      | 'meldextra'
      | 'layoff'
      | 'discard'
      | 'nextround'
      | 'log',
    params?: {
      cardIndex?: number;
      cardIndices?: number[];
      indicesPerSlot?: number[][];
      targetPlayerIdx?: number;
      meldIdx?: number;
      config?: { playerCount?: number; cpuDifficulty?: number; failContractPenalty?: number };
    },
  ) => gameExec<CariocaResponse>('carioca', { command, ...(params ?? {}) }),
};
