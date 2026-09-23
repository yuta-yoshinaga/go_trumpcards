// API client for tapptarock. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TappTarockResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the TappTarock /tapptarock/exec endpoint.
 *
 * `bid` takes `"dreier"` or `"solo"` — **never `"trischaken"`**,
 * which only arises when everyone passes and so cannot be declared.
 * `cardIndices` are the six 0-based hand positions to bury.
 */
export const tapptarockApi = {
  exec: (
    command: 'reset' | 'bid' | 'pass' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    params?: {
      bid?: 'dreier' | 'solo';
      cardIndex?: number;
      cardIndices?: number[];
      config?: { cpuDifficulty?: number; targetDeals?: number };
    },
  ) => gameExec<TappTarockResponse>('tapptarock', { command, ...params }),
};
