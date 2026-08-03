// API client for pontoon. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PontoonResponse } from '../../types/card';
import { createBetAmountApi } from '../gameExec';

/** Commands accepted by the /pontoon/exec endpoint. */
export type PontoonCommand =
  | 'reset'
  | 'bet'
  | 'deal'
  | 'stick'
  | 'twist'
  | 'buy'
  | 'split'
  | 'bankertwist'
  | 'bankerstay'
  | 'log';

/**
 * API client for the Pontoon /pontoon/exec endpoint.
 *
 * `bet` and `buy` are the only commands that carry an amount, so the shared
 * bet-amount factory fits: every other command simply omits it.
 */
export const pontoonApi = createBetAmountApi<PontoonResponse, PontoonCommand>('pontoon');
