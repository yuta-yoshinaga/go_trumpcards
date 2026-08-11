// API client for bhabhi. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BhabhiConfig, BhabhiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Bhabhi /bhabhi/exec endpoint.
 *
 * **There is no `next` command.** The whole deck is dealt once and play runs
 * until only one player still holds cards, so the game has no hand boundaries.
 */
export const bhabhiApi = {
  exec: (command: 'reset' | 'play' | 'giveup' | 'hint' | 'log', cardIndex?: number, config?: Partial<BhabhiConfig>) =>
    gameExec<BhabhiResponse>('bhabhi', { command, cardIndex, config }),
};
