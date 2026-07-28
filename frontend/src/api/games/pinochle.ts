// API client for pinochle. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PinochleResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Pinochle game settings. */
export interface PinochleConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Pinochle /pinochle/exec endpoint. */
export const pinochleApi = {
  exec: (
    command: 'reset' | 'bid' | 'pass' | 'trump' | 'meld' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndex?: number,
    config?: PinochleConfigInput,
    bidAmount?: number,
    suit?: number,
  ) =>
    gameExec<PinochleResponse>('pinochle', {
      command,
      cardIndex,
      config,
      bidAmount,
      suit,
    }),
};
