// API client for reddog. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RedDogResponse } from '../../types/card';
import { createBetAmountApi } from '../gameExec';

/** API client for the Red Dog /reddog/exec endpoint. */
export const reddogApi = createBetAmountApi<RedDogResponse, 'reset' | 'bet' | 'raise' | 'stay' | 'log' | 'hint'>(
  'reddog',
);
