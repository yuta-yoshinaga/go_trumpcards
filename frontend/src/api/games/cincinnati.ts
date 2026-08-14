// API client for cincinnati. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CincinnatiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Cincinnati /cincinnati/exec endpoint.
 *
 * `amount` is required for `bet` and `raise` and ignored by the other actions,
 * so the caller may omit it entirely when folding, checking or calling.
 */
export const cincinnatiApi = {
  exec: (
    command: 'reset' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'next' | 'hint' | 'log',
    params?: { amount?: number },
  ) => gameExec<CincinnatiResponse>('cincinnati', { command, ...params }),
};
