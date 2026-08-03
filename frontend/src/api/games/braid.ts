// API client for braid. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BraidMoveZone, BraidResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands accepted by the /braid/exec endpoint. */
export type BraidCommand =
  | 'reset'
  | 'draw'
  | 'dir'
  | 'move'
  | 'giveup'
  | 'hint'
  | 'autocomplete'
  | 'log'
  | 'undo'
  | 'undo_n';

/**
 * API client for the Braid /braid/exec endpoint.
 *
 * Braid does not use {@link createSolitaireMoveApi} because `dir` carries an
 * `ascending` flag that no other solitaire command needs.
 */
export const braidApi = {
  exec: (command: BraidCommand, from?: BraidMoveZone, to?: BraidMoveZone, n?: number, ascending?: boolean) =>
    gameExec<BraidResponse>('braid', { command, from, to, n, ascending }),
};
