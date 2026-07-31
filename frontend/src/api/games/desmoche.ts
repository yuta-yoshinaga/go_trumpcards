// API client for desmoche. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DesmocheResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /desmoche/exec endpoint accepts. */
export type DesmocheCommand =
  | 'reset'
  | 'drawstock'
  | 'drawdiscard'
  | 'meld'
  | 'layoff'
  | 'desmoche'
  | 'discard'
  | 'next'
  | 'hint'
  | 'log';

/** Extra indices the `desmoche` (rearrange) command needs. */
export interface DesmocheMoveIndices {
  fromMeldIndex: number;
  toMeldIndex: number;
}

/**
 * API client for the Desmoche /desmoche/exec endpoint.
 *
 * `cardIndex` and `meldIndex` are separate parameters so one can never be read
 * as the other. For the `desmoche` command `cardIndex` is a position *within*
 * `fromMeldIndex`, not a hand position.
 */
export const desmocheApi = {
  exec: (
    command: DesmocheCommand,
    cardIndex?: number,
    meldIndex?: number,
    cardIndices?: number[],
    move?: DesmocheMoveIndices,
  ) =>
    gameExec<DesmocheResponse>('desmoche', {
      command,
      cardIndex,
      meldIndex,
      cardIndices,
      fromMeldIndex: move?.fromMeldIndex,
      toMeldIndex: move?.toMeldIndex,
    }),
};
