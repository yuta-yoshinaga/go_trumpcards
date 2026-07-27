// API client for pineapple. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PineappleResponse } from '../../types/card';
import { gameExec } from '../gameExec';
import type { HoldemConfigInput, HoldemLikeCommand } from './holdem';

/** Configuration options for Pineapple Poker (extends Hold'em with cardIdx/cardIdxs for discard). */
export interface PineappleConfigInput extends HoldemConfigInput {
  cardIdx?: number;
  /** Multiple discard indices, submitted together (Irish Poker's 2-card discard). */
  cardIdxs?: number[];
}

/** API client for the Pineapple Poker /pineapple/exec endpoint. */
export const pineappleApi = {
  exec: (
    command: HoldemLikeCommand | 'discard',
    amount?: number,
    config?: PineappleConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<PineappleResponse>('pineapple', {
      command,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};
