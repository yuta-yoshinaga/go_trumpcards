// API client for doubleklondike. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DoubleKlondikeResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands accepted by the Double Klondike /doubleklondike/exec endpoint. */
export type DoubleKlondikeCommand =
  | 'reset'
  | 'd'
  | 'mwt'
  | 'mwf'
  | 'mtt'
  | 'mtf'
  | 'g'
  | 'ac'
  | 'u'
  | 'undo_n'
  | 'hint'
  | 'log';

/** API client for the Double Klondike /doubleklondike/exec endpoint. */
export const doubleklondikeApi = {
  exec: (
    command: DoubleKlondikeCommand,
    opts?: { col?: number; fromCol?: number; cardIndex?: number; toCol?: number; n?: number },
  ) =>
    gameExec<DoubleKlondikeResponse>('doubleklondike', {
      command,
      col: opts?.col,
      fromCol: opts?.fromCol,
      cardIndex: opts?.cardIndex,
      toCol: opts?.toCol,
      n: opts?.n,
    }),
};
