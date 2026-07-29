// API client for settemezzo. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SetteEMezzoResponse } from '../../types/card';
import { createBetAmountApi } from '../gameExec';

/** Commands accepted by the /settemezzo/exec endpoint. */
export type SetteEMezzoCommand =
  | 'reset'
  | 'bet'
  | 'deal'
  | 'hit'
  | 'stand'
  | 'matta'
  | 'bankerhit'
  | 'bankerstand'
  | 'log';

/**
 * API client for the Sette e Mezzo /settemezzo/exec endpoint.
 *
 * `bet` carries a stake and `matta` carries the wild card's value in HALVES;
 * every other command omits the amount.
 */
export const settemezzoApi = createBetAmountApi<SetteEMezzoResponse, SetteEMezzoCommand>('settemezzo');
