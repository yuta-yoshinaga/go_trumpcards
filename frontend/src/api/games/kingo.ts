// API client for kingo. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KingoResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Kingo /kingo/exec endpoint.
 *
 * `amount` is required for `bet` and ignored by the other actions. Betting and
 * dealing are separate commands because the human owes a different action
 * depending on whether the bank is theirs this round.
 */
export const kingoApi = {
  exec: (command: 'reset' | 'bet' | 'deal' | 'next' | 'hint' | 'log', params?: { amount?: number }) =>
    gameExec<KingoResponse>('kingo', { command, ...params }),
};
