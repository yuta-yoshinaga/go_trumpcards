// API client for ohhell. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OhHellResponse } from '../../types/card';
import { createBidPlayApi } from '../gameExec';

/** Configuration options for Oh Hell game settings. */
export interface OhHellConfigInput {
  cpuDifficulty?: number;
  maxHandSize?: number;
  scoringVariant?: number;
  roundDirection?: number;
}

/** API client for the Oh Hell /ohhell/exec endpoint. */
export const ohHellApi = createBidPlayApi<OhHellResponse, OhHellConfigInput>('ohhell');
