// API client for montecarlo. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MonteCarloResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Monte Carlo Solitaire /montecarlo/exec endpoint. */
export const montecarloApi = {
  exec: (
    command: 'reset' | 'remove' | 'deal' | 'undo' | 'giveup' | 'hint' | 'log',
    fromR?: number,
    fromC?: number,
    toR?: number,
    toC?: number,
  ) => gameExec<MonteCarloResponse>('montecarlo', { command, fromR, fromC, toR, toC }),
};
