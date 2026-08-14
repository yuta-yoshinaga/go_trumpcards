// API client for zwanzigerrufen. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ZwanzigerrufenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Zwanzigerrufen /zwanzigerrufen/exec endpoint.
 *
 * `bid` takes `"rufer"` (call the XX) or `"solo"` — **never `"trischaken"`**,
 * which only arises when everyone passes and so cannot be declared.
 * `cardIndices` are the six 0-based hand positions to bury.
 */
export const zwanzigerrufenApi = {
  exec: (
    command: 'reset' | 'bid' | 'pass' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    params?: {
      bid?: 'rufer' | 'solo';
      cardIndex?: number;
      cardIndices?: number[];
      config?: { cpuDifficulty?: number; targetDeals?: number };
    },
  ) => gameExec<ZwanzigerrufenResponse>('zwanzigerrufen', { command, ...params }),
};
