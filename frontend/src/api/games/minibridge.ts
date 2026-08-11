// API client for minibridge. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MinibridgeConfig, MinibridgeResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Minibridge /minibridge/exec endpoint.
 *
 * **There is no bid command.** The declarer is decided from the HCP everyone
 * announces, and the only declaration is the contract — `contract` needs both
 * level and suit, and `suit: 0` means no-trump rather than "omitted".
 */
export const minibridgeApi = {
  exec: (
    command: 'reset' | 'contract' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<MinibridgeConfig>,
    level?: number,
    suit?: number,
  ) => gameExec<MinibridgeResponse>('minibridge', { command, cardIndex, config, level, suit }),
};
