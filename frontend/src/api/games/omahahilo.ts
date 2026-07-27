// API client for omahahilo. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OmahaResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';

/** API client for the Omaha Hi-Lo / 8 or Better /omahahilo/exec endpoint.
 * Shares the OmahaResponse shape — additional Hi-Lo fields (LowBestHand,
 * LowQualifies, HiWonAmount, LowWonAmount) are surfaced via omitempty
 * JSON encoding so the same TypeScript type works for both. */
export const omahaHiLoApi = createHoldemLikeApi<OmahaResponse>('omahahilo');
