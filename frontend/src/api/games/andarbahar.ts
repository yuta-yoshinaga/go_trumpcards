// API client for andarbahar. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { AndarBaharResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Andar Bahar /andarbahar/exec endpoint.
 *
 * **One `bet` resolves the whole round.** Cards are dealt alternately until one
 * matches the joker, with no decision in between.
 *
 * `sideBand` is only read when `sideAmount` is above 0 — omitting the side bet
 * must not be mistaken for staking 0 on band 0.
 */
export const andarbaharApi = {
  exec: (
    command: 'reset' | 'bet' | 'clear' | 'hint' | 'log',
    amount?: number,
    target?: number,
    sideAmount?: number,
    sideBand?: number,
  ) =>
    gameExec<AndarBaharResponse>('andarbahar', {
      command,
      amount,
      target,
      sideAmount,
      sideBand,
    }),
};
