// API client for letitride. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { LetItRideResponse } from '../../types/card';
import { createBetAmountApi } from '../gameExec';

/** API client for the Let It Ride /letitride/exec endpoint. */
export const letitrideApi = createBetAmountApi<LetItRideResponse, 'reset' | 'bet' | 'pull' | 'letitride' | 'log'>(
  'letitride',
);
