// API client for quinze. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { QuinzeResponse } from '../../types/card';
import { createBetAmountApi } from '../gameExec';

/** Commands accepted by the /quinze/exec endpoint. */
export type QuinzeCommand = 'reset' | 'bet' | 'deal' | 'hit' | 'stand' | 'bankerhit' | 'bankerstand' | 'log';

/**
 * API client for the Quinze /quinze/exec endpoint.
 *
 * `bet` carries a stake; every other command omits the amount.
 */
export const quinzeApi = createBetAmountApi<QuinzeResponse, QuinzeCommand>('quinze');
