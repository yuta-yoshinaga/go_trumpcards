// API client for callbreak. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CallBreakResponse } from '../../types/card';
import { createBidPlayApi } from '../gameExec';

/** Configuration options for Call Break game settings. */
export interface CallBreakConfigInput {
  cpuDifficulty?: number;
  maxRounds?: number;
}

/** API client for the Call Break /callbreak/exec endpoint. */
export const callBreakApi = createBidPlayApi<CallBreakResponse, CallBreakConfigInput>('callbreak');
