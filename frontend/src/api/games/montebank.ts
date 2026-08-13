// API client for montebank. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MonteBankResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Monte Bank /montebank/exec endpoint.
 *
 * `idx` is **0-based and 0 is a valid value** (the leftmost layout card), so it
 * must always be sent explicitly — the server rejects a `bet` with no index
 * rather than treating it as "the first card".
 */
export const montebankApi = {
  exec: (command: 'reset' | 'bet' | 'next' | 'hint' | 'log', params?: { idx?: number; bet?: number }) =>
    gameExec<MonteBankResponse>('montebank', { command, ...params }),
};
