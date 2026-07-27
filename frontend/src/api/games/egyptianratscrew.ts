// API client for egyptianratscrew. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { EgyptianRatscrewResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Egyptian Ratscrew game settings. */
export interface EgyptianRatscrewConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Egyptian Ratscrew /egyptianratscrew/exec endpoint. */
export const egyptianRatscrewApi = {
  exec: (command: 'reset' | 'step' | 'slap' | 'tick' | 'log', args?: { config?: EgyptianRatscrewConfigInput }) =>
    gameExec<EgyptianRatscrewResponse>('egyptianratscrew', {
      command,
      ...(args || {}),
    }),
};
