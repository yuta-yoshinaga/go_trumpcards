import type { BassetResponse } from '../../types/games/basset';
import { gameExec } from '../gameExec';

/** Commands accepted by the Basset endpoint. */
export type BassetCommand = 'reset' | 'bet' | 'deal' | 'take' | 'paroli' | 'next' | 'log';

/** Calls the Basset endpoint. */
export const bassetApi = {
  exec: (command: BassetCommand, opts?: { rank?: number; amount?: number }) =>
    gameExec<BassetResponse>('basset', { command, rank: opts?.rank, amount: opts?.amount }),
};
