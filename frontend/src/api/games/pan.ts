// API client for pan. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PanResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Panguingue (Pan) game settings. */
export interface PanConfigInput {
  playerCount?: number;
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Action parameters for a Panguingue (Pan) turn. */
export interface PanActionParams {
  /** Hand-card indices forming a new meld (set or rope). */
  cardIndices?: number[];
  /** Hand-card index to discard or lay off. */
  cardIndex?: number;
  /** Owning player id of the target meld for a layoff. */
  meldOwner?: number;
  /** Index of the target meld within the owner's laid melds for a layoff. */
  meldIdx?: number;
}

/** API client for the Panguingue (Pan) /pan/exec endpoint. */
export const panApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'layoff' | 'discard' | 'nextround' | 'log',
    params?: PanActionParams,
    config?: PanConfigInput,
  ) =>
    gameExec<PanResponse>('pan', {
      command,
      cardIndices: params?.cardIndices,
      cardIndex: params?.cardIndex,
      meldOwner: params?.meldOwner,
      meldIdx: params?.meldIdx,
      config,
    }),
};
