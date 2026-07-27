// API client for mighty. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MightyResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Mighty game settings. */
export interface MightyConfigInput {
  cpuDifficulty?: number;
  minBid?: number;
  noTrumpExtra?: number;
  pointLimit?: number;
}

/** API client for the Mighty /mighty/exec endpoint. */
export const mightyApi = {
  exec: (
    command:
      | 'reset'
      | 'b'
      | 'bid'
      | 't'
      | 'trump'
      | 'e'
      | 'exchange'
      | 'p'
      | 'play'
      | 'jl'
      | 'jokerlead'
      | 'n'
      | 'next'
      | 'nr'
      | 'nextround'
      | 'hint'
      | 'log',
    bid?: number,
    noTrump?: boolean,
    cardIndex?: number,
    trumpSuit?: number,
    partnerSuit?: number,
    partnerValue?: number,
    discardIndices?: number[],
    jokerLeadSuit?: number,
    config?: MightyConfigInput,
  ) =>
    gameExec<MightyResponse>('mighty', {
      command,
      bid,
      noTrump,
      cardIndex,
      trumpSuit,
      partnerSuit,
      partnerValue,
      discardIndices,
      jokerLeadSuit,
      config,
    }),
};
