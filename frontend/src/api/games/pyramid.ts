// API client for pyramid. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PyramidResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Source card for a Pyramid remove action. */
export interface PyramidRemoveCard {
  zone: string;
  row?: number;
  col?: number;
}

/** API client for the Pyramid /pyramid/exec endpoint. */
export const pyramidApi = {
  exec: (
    command: 'reset' | 'draw' | 'remove' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n',
    card1?: PyramidRemoveCard,
    card2?: PyramidRemoveCard,
    n?: number,
  ) => gameExec<PyramidResponse>('pyramid', { command, card1, card2, n }),
};
