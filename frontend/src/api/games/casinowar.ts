// API client for casinowar. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CasinoWarResponse } from '../../types/card';
import { createBetAmountApi } from '../gameExec';

/** API client for the Casino War /casinowar/exec endpoint. */
export const casinowarApi = createBetAmountApi<CasinoWarResponse, 'reset' | 'bet' | 'surrender' | 'war' | 'log'>(
  'casinowar',
);
