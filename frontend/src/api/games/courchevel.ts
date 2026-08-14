// API client for courchevel. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OmahaResponse } from '../../types/card';
import { createHoldemLikeApi } from './holdem';

/**
 * API client for the Courchevel /courchevel/exec endpoint.
 *
 * Shares the OmahaResponse shape — Courchevel is Big O with the first flop
 * card turned before the opening bet, so nothing about the wire changes:
 * `communityCards` simply already holds one card during the pre-flop phase.
 */
export const courchevelApi = createHoldemLikeApi<OmahaResponse>('courchevel');
