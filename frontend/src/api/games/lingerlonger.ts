// API client for lingerlonger. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { LingerLongerConfig, LingerLongerResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Linger Longer /lingerlonger/exec endpoint.
 *
 * **There is no draw command.** Winning a trick refills your hand
 * automatically, so playing a card is the only move the game accepts.
 */
export const lingerlongerApi = {
  exec: (
    command: 'reset' | 'play' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<LingerLongerConfig>,
  ) => gameExec<LingerLongerResponse>('lingerlonger', { command, cardIndex, config }),
};
