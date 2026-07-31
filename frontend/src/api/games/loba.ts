// API client for loba. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { LobaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /loba/exec endpoint accepts. */
export type LobaCommand =
  | 'reset'
  | 'drawstock'
  | 'drawdiscard'
  | 'meld'
  | 'layoff'
  | 'discard'
  | 'next'
  | 'hint'
  | 'log';

/**
 * API client for the Loba /loba/exec endpoint.
 *
 * `cardIndex` (a hand position) and `meldIndex` (a table position) are separate
 * parameters so one can never be read as the other.
 */
export const lobaApi = {
  exec: (command: LobaCommand, cardIndex?: number, meldIndex?: number, cardIndices?: number[]) =>
    gameExec<LobaResponse>('loba', { command, cardIndex, meldIndex, cardIndices }),
};
