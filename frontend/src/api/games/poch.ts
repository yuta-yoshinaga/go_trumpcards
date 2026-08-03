// API client for poch. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PochResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /poch/exec endpoint accepts. */
export type PochCommand = 'reset' | 'bet' | 'fold' | 'play' | 'next' | 'hint' | 'log';

/** API client for the Poch /poch/exec endpoint. */
export const pochApi = {
  exec: (command: PochCommand, cardIndex?: number) => gameExec<PochResponse>('poch', { command, cardIndex }),
};
