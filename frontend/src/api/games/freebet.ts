// API client for freebet. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FreeBetResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Free Bet Blackjack /freebet/exec endpoint.
 *
 * `freedouble` and `freesplit` take no parameters — the house pays the raise,
 * so there is no amount to send. The server decides whether either is legal;
 * a rejected command comes back as a message rather than an exception.
 */
export const freebetApi = {
  exec: (
    command: 'reset' | 'bet' | 'hit' | 'stand' | 'freedouble' | 'freesplit' | 'next' | 'hint' | 'log',
    params?: { ante?: number },
  ) => gameExec<FreeBetResponse>('freebet', { command, ...params }),
};
