// API client for bigo. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OmahaResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';

/** API client for the 5 Card Omaha (Big O) /bigo/exec endpoint.
 * Shares the OmahaResponse shape — Big O is Omaha dealt 5 hole cards. */
export const bigOApi = createHoldemLikeApi<OmahaResponse>('bigo');
