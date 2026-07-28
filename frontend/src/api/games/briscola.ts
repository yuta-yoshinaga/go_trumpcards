// API client for briscola. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BriscolaConfig, BriscolaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Briscola /briscola/exec endpoint. */
export const briscolaApi = {
  exec: (command: 'reset' | 'play' | 'next' | 'hint' | 'log', cardIndex?: number, config?: Partial<BriscolaConfig>) =>
    gameExec<BriscolaResponse>('briscola', { command, cardIndex, config }),
};
