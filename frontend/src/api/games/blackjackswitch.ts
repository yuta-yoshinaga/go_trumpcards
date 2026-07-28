// API client for blackjackswitch. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BlackJackSwitchResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Blackjack Switch /blackjackswitch/exec endpoint. */
export const blackjackswitchApi = {
  exec: (command: 'reset' | 'bet' | 'switch' | 'keep' | 'hit' | 'stand' | 'doubledown' | 'log', amount?: number) =>
    gameExec<BlackJackSwitchResponse>('blackjackswitch', { command, amount }),
};
