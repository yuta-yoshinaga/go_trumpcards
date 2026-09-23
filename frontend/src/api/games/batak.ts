// API client for batak. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BatakResponse } from '../../types/card';
import { createBidPlayApi } from '../gameExec';

/** Configuration options for Batak game settings. */
export interface BatakConfigInput {
  cpuDifficulty?: number;
  maxRounds?: number;
}

/** API client for the Batak /batak/exec endpoint. */
export const batakApi = createBidPlayApi<BatakResponse, BatakConfigInput>('batak');
