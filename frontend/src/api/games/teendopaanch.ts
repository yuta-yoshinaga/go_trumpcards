// API client for teendopaanch. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TeenDoPaanchConfig, TeenDoPaanchResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the 3-2-5 /teendopaanch/exec endpoint.
 *
 * **There is no bid command.** The 3, 2 and 5 targets are assigned at the start
 * of each round and rotate; the only declaration is trump, and only the seat
 * that owes 5 tricks makes it.
 */
export const teendopaanchApi = {
  exec: (
    command: 'reset' | 'trump' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<TeenDoPaanchConfig>,
    suit?: number,
  ) => gameExec<TeenDoPaanchResponse>('teendopaanch', { command, cardIndex, config, suit }),
};
