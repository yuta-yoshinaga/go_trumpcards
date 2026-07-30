// API client for bura. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BuraResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /bura/exec endpoint accepts. */
export type BuraCommand = 'reset' | 'play' | 'claim' | 'declare' | 'hint' | 'log';

/**
 * API client for the Bura /bura/exec endpoint.
 *
 * `cardIndices` is a LIST, not one index: a Bura lead may be up to three cards
 * of a single suit, and a response has to match that count.
 */
export const buraApi = {
  exec: (command: BuraCommand, cardIndices?: number[]) => gameExec<BuraResponse>('bura', { command, cardIndices }),
};
