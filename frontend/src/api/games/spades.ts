// API client for spades. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SpadesResponse } from '../../types/card';
import { createBidPlayApi } from '../gameExec';

/** Configuration options for Spades game settings. */
export interface SpadesConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
  nilBonus?: number;
  bagPenaltyThreshold?: number;
}

/** API client for the Spades /spades/exec endpoint. */
export const spadesApi = createBidPlayApi<SpadesResponse, SpadesConfigInput>('spades');
