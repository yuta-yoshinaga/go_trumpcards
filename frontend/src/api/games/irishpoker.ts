// API client for irishpoker. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PineappleResponse } from '../../types/card';
import { gameExec } from '../gameExec';
import type { HoldemLikeCommand } from './holdem';
import type { PineappleConfigInput } from './pineapple';

/** API client for the Irish Poker /irishpoker/exec endpoint. */
export const irishPokerApi = {
  exec: (
    command: HoldemLikeCommand | 'discard',
    amount?: number,
    config?: PineappleConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<PineappleResponse>('irishpoker', {
      command,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};
