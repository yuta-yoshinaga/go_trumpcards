// API client for sakura. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SakuraResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Sakura /sakura/exec endpoint.
 *
 * `cardIndex` is a 0-based hand position. `fieldIndex` only matters when two
 * field cards share the played card's month — leave it out and the server takes
 * the higher-scoring one.
 */
export const sakuraApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'hint' | 'log',
    params?: { cardIndex?: number; fieldIndex?: number; config?: Partial<SakuraResponse['config']> },
  ) => gameExec<SakuraResponse>('sakura', { command, ...params }),
};
