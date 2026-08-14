// API client for mendikot. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MendikotConfig, MendikotResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Mendikot /mendikot/exec endpoint.
 *
 * **There is no trump command.** The suit is fixed by whichever card the first
 * player who cannot follow chooses to play, so it arrives through `play`.
 */
export const mendikotApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<MendikotConfig>,
  ) => gameExec<MendikotResponse>('mendikot', { command, cardIndex, config }),
};
