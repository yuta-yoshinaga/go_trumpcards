// API client for contractrummy. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ContractRummyResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Contract Rummy /contractrummy/exec endpoint. */
export const contractrummyApi = {
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
      config?: { cpuDifficulty?: number; failContractPenalty?: number };
    },
  ) => gameExec<ContractRummyResponse>('contractrummy', { command, ...(params ?? {}) }),
};
