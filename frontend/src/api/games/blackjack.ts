// API client for blackjack. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BlackJackResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for BlackJack game settings. */
export interface BlackJackConfigInput {
  dealerHitsSoft17?: boolean;
  cpuPlayerCount?: number;
  countingEnabled?: boolean;
  doubleAfterSplit?: boolean;
  countingSystem?: number;
  deckPenetration?: number;
  surrenderRule?: number;
}

/** Side bet and multi-hand options for BlackJack. */
export interface BlackJackBetOptions {
  perfectPairsBet?: number;
  twentyOnePlus3Bet?: number;
  handCount?: number;
}

/** Type alias for BlackJack/Spanish21 exec command. */
export type BlackJackCommand =
  | 'reset'
  | 'hit'
  | 'stand'
  | 'bet'
  | 'doubledown'
  | 'split'
  | 'insurance'
  | 'declineinsurance'
  | 'surrender'
  | 'togglehint'
  | 'setdeckcount'
  | 'togglesoft17'
  | 'togglecounting'
  | 'toggledas'
  | 'setcountingsystem'
  | 'setpenetration'
  | 'setcpucount'
  | 'earlysurrender'
  | 'declineearlysurrender'
  | 'setsurrenderrule';

/**
 * Factory for BlackJack-shaped APIs whose request body is
 * `{ command, amount, ...config, ...betOptions }`. Spanish 21 reuses the
 * BlackJack response and command union, so both clients are constructed
 * via the same factory rather than duplicating the eight-token shape.
 *
 * Kept narrow on purpose — only games that share the BlackJack command
 * union and bet-option payload should use it. See issue #1550.
 */
export function createBlackJackLikeApi(game: string) {
  return {
    exec: (
      command: BlackJackCommand,
      amount?: number,
      config?: BlackJackConfigInput,
      betOptions?: BlackJackBetOptions,
    ) => gameExec<BlackJackResponse>(game, { command, amount, ...config, ...betOptions }),
  };
}

/** API client for the BlackJack /blackjack/exec endpoint. */
export const blackjackApi = createBlackJackLikeApi('blackjack');
