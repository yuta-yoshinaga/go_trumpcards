// API client for sergeantmajor. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SergeantMajorConfig, SergeantMajorResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Sergeant Major /sergeantmajor/exec endpoint.
 *
 * **There is no bid command.** The 8, 5 and 3 targets are fixed by seat and
 * rotate with the deal; the only declaration is trump, and only the dealer
 * makes it.
 */
export const sergeantmajorApi = {
  exec: (
    command: 'reset' | 'trump' | 'discard' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<SergeantMajorConfig>,
    suit?: number,
    discards?: number[],
  ) => gameExec<SergeantMajorResponse>('sergeantmajor', { command, cardIndex, config, suit, discards }),
};
