// API client for bigohilo. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OmahaResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';

/** API client for the 5 Card Omaha Hi-Lo (Big O) /bigohilo/exec endpoint.
 * Shares the OmahaResponse shape (Hi-Lo split fields surfaced via omitempty). */
export const bigOHiLoApi = createHoldemLikeApi<OmahaResponse>('bigohilo');
