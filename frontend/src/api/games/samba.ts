// API client for samba. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SambaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Samba game settings. */
export interface SambaConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Samba /samba/exec endpoint. */
export const sambaApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'skipmeld' | 'discard' | 'goout' | 'nextround' | 'log',
    cardIndex?: number,
    config?: SambaConfigInput,
    naturalPairIndices?: number[],
    meldGroups?: number[][],
  ) =>
    gameExec<SambaResponse>('samba', {
      command,
      cardIndex,
      config,
      naturalPairIndices,
      meldGroups,
    }),
};
