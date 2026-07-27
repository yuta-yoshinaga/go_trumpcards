// API client for war. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { WarResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the War /war/exec endpoint. */
export const warApi = {
  exec: (command: 'reset' | 'step' | 'autoplay' | 'log', config?: { maxRounds?: number }) =>
    gameExec<WarResponse>('war', { command, ...config }),
};
