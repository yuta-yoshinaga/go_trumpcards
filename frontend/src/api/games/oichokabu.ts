// API client for oichokabu. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OichoKabuResponse } from '../../types/card';
import { createBetAmountApi } from '../gameExec';

/** API client for the Oicho-Kabu /oichokabu/exec endpoint. */
export const oichokabuApi = createBetAmountApi<OichoKabuResponse, 'reset' | 'bet' | 'draw' | 'stand' | 'log'>(
  'oichokabu',
);
