// API client for handandfoot. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { HandAndFootResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Hand and Foot game settings. */
export interface HandAndFootConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Hand and Foot /handandfoot/exec endpoint. */
export const handandfootApi = {
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
    config?: HandAndFootConfigInput,
    naturalPairIndices?: number[],
    meldGroups?: number[][],
  ) =>
    gameExec<HandAndFootResponse>('handandfoot', {
      command,
      cardIndex,
      config,
      naturalPairIndices,
      meldGroups,
    }),
};
