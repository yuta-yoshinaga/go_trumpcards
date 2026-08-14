// API client for rollingstone. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RollingStoneConfig, RollingStoneResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Rolling Stone /rollingstone/exec endpoint.
 *
 * **`pickup` is its own command, not `play` with the index omitted.** Being
 * unable to follow is a different action, not a missing parameter.
 */
export const rollingstoneApi = {
  exec: (
    command: 'reset' | 'play' | 'pickup' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<RollingStoneConfig>,
  ) => gameExec<RollingStoneResponse>('rollingstone', { command, cardIndex, config }),
};
