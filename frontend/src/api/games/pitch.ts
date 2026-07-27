// API client for pitch. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PitchResponse } from '../../types/card';
import { createBidPlayApi } from '../gameExec';

/** Configuration options for Pitch (Setback) game settings. */
export interface PitchConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Pitch /pitch/exec endpoint. */
export const pitchApi = createBidPlayApi<PitchResponse, PitchConfigInput>('pitch');
