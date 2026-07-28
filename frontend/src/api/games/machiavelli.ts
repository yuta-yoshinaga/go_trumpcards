// API client for machiavelli. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MachiavelliResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Move parameters for a Machiavelli turn action (newmeld / layoff / play). */
export interface MachiavelliMoveParams {
  /** Full proposed table for the `play` power move (card refs by design + value). */
  tableMelds?: { design: number; value: number }[][];
  /** Hand-card indices for `newmeld` (or the cards added by `play`). */
  handIndices?: number[];
  /** Target table meld index for `layoff`. */
  meldIdx?: number;
  /** Hand-card index added to an existing meld for `layoff`. */
  handIndex?: number;
}

/** Configuration options for Machiavelli game settings. */
export interface MachiavelliConfigInput {
  playerCount?: number;
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** API client for the Machiavelli /machiavelli/exec endpoint. */
export const machiavelliApi = {
  exec: (
    command: 'reset' | 'draw' | 'play' | 'newmeld' | 'layoff' | 'nextround' | 'log',
    params?: MachiavelliMoveParams,
    config?: MachiavelliConfigInput,
  ) =>
    gameExec<MachiavelliResponse>('machiavelli', {
      command,
      ...(params ?? {}),
      config,
    }),
};
