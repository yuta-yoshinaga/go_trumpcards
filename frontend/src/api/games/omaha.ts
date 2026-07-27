// API client for omaha. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OmahaResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';

/** API client for the Omaha Hold'em /omaha/exec endpoint. */
export const omahaApi = createHoldemLikeApi<OmahaResponse>('omaha');
