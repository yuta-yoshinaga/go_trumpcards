// API client for troggu. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TrogguResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Troggu /troggu/exec endpoint.
 *
 * `bid` takes one of the four contracts. **There is no "pass" contract**: pass
 * is its own command, and a deal where everyone passes is thrown in.
 */
export const trogguApi = {
  exec: (
    command: 'reset' | 'bid' | 'pass' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    params?: {
      bid?: 'trois' | 'solo' | 'piccolo' | 'misere';
      cardIndex?: number;
      config?: { cpuDifficulty?: number; targetDeals?: number };
    },
  ) => gameExec<TrogguResponse>('troggu', { command, ...params }),
};
