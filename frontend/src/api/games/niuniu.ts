// API client for niuniu. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { NiuNiuResponse } from '../../types/card';
import { createBetAmountApi } from '../gameExec';

/**
 * API client for the Niu Niu /niuniu/exec endpoint.
 *
 * There is no turn in this game, so `bet` is the only action -- it deals and
 * settles in one call.
 */
export const niuniuApi = createBetAmountApi<NiuNiuResponse, 'reset' | 'bet' | 'log'>('niuniu');
