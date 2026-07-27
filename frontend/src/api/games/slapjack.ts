// API client for slapjack. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SlapjackResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Slapjack game settings. */
export interface SlapjackConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Slapjack /slapjack/exec endpoint. */
export const slapjackApi = {
  exec: (command: 'reset' | 'step' | 'slap' | 'tick' | 'log', args?: { config?: SlapjackConfigInput }) =>
    gameExec<SlapjackResponse>('slapjack', {
      command,
      ...(args || {}),
    }),
};
