// API client for eightgame. Follows the split-out convention of gameApi.ts
// (issue #4434); gameApi.ts re-exports this file.

import type { HorseResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Eight-Game Mix /eightgame/exec endpoint.
 *
 * Same orchestrator as H.O.R.S.E., so the response type and the verbs are
 * shared — the rotation is eight disciplines instead of five, and `draw`
 * exchanges cards during the 2-7 Triple Draw hands.
 */
export const eightGameApi = {
  exec: (
    command: 'reset' | 'action' | 'draw' | 'next' | 'hint' | 'log',
    params?: {
      action?: 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin';
      amount?: number;
      /** Cards to exchange in a draw round, 0-based. Omitted = stand pat. */
      cardIndices?: number[];
      config?: { seats?: number; initialChips?: number; handsPerDiscipline?: number };
    },
  ) => gameExec<HorseResponse>('eightgame', { command, ...params }),
};
