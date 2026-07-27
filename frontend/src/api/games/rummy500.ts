// API client for rummy500. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { Rummy500Response } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Rummy 500 game settings. */
export interface Rummy500ConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** Layoff parameters: meld owner, meld index, and card index in hand. */
export interface Rummy500LayoffInput {
  meldOwner: number;
  meldIdx: number;
  cardIndex: number;
}

/** API client for the Rummy 500 /rummy500/exec endpoint. */
export const rummy500Api = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'layoff' | 'discard' | 'nextround' | 'log',
    cardIndex?: number,
    config?: Rummy500ConfigInput,
    cardIndices?: number[],
    discardIdx?: number,
    layoff?: Rummy500LayoffInput,
  ) =>
    gameExec<Rummy500Response>('rummy500', {
      command,
      cardIndex: layoff?.cardIndex ?? cardIndex,
      cardIndices,
      discardIdx,
      meldOwner: layoff?.meldOwner,
      meldIdx: layoff?.meldIdx,
      config,
    }),
};
