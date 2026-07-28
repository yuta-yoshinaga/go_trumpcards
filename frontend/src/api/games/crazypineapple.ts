// API client for crazypineapple. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PineappleResponse } from '../../types/card';
import { gameExec } from '../gameExec';
import type { HoldemLikeCommand } from './holdem';
import type { PineappleConfigInput } from './pineapple';

/** API client for the Crazy Pineapple Poker /crazypineapple/exec endpoint. */
export const crazyPineappleApi = {
  exec: (
    command: HoldemLikeCommand | 'discard',
    amount?: number,
    config?: PineappleConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<PineappleResponse>('crazypineapple', {
      command,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};
