// API client for courchevelhilo. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OmahaResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';

/** API client for the Courchevel Hi-Lo /courchevelhilo/exec endpoint.
 * Shares the OmahaResponse shape (Hi-Lo split fields surfaced via omitempty). */
export const courchevelHiLoApi = createHoldemLikeApi<OmahaResponse>('courchevelhilo');
