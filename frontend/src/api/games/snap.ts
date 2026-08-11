// API client for snap. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SnapResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Config accepted by Snap's reset. */
export interface SnapConfigInput {
  playerCnt?: number;
  cpuDifficulty?: number;
}

/**
 * API client for the Snap /snap/exec endpoint.
 *
 * **`snap` takes no seat.** The server always calls for seat 0 — letting a
 * client name the seat would let it force a CPU into a wrong call.
 */
export const snapApi = {
  exec: (command: 'reset' | 'step' | 'snap' | 'tick' | 'giveup' | 'hint' | 'log', config?: SnapConfigInput) =>
    gameExec<SnapResponse>('snap', { command, config }),
};
