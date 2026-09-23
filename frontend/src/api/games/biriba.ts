// API client for biriba. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BiribaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Biriba game settings. */
export interface BiribaConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Biriba /biriba/exec endpoint. */
export const biribaApi = {
  exec: (
    command:
      | 'reset'
      | 'drawstock'
      | 'drawdiscard'
      | 'meld'
      | 'skipmeld'
      | 'discard'
      | 'goout'
      | 'nextround'
      | 'log'
      | 'hint',
    cardIndex?: number,
    config?: BiribaConfigInput,
    naturalPairIndices?: number[],
    meldGroups?: number[][],
  ) =>
    gameExec<BiribaResponse>('biriba', {
      command,
      cardIndex,
      config,
      naturalPairIndices,
      meldGroups,
    }),
};
